package flatfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"
	"time"

	"os"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

type FlatFile interface {
	//append item into flat file, return item index
	Append([]byte) (uint64, error)

	//read item in flatfile with specific index (return from append)
	Read(index uint64) ([]byte, error)

	//read recent data, return data channel, errpr channel, and cancel function
	ReadRecently(index uint64) (chan []byte, chan int, func())

	//truncate flat file system
	Truncate(lastIndex uint64) error
}

type FlatFileManager struct {
	dataDir         string
	fileSizeLimit   uint64 //number of item
	folderMap       map[uint64]bool
	sortedFolder    []uint64
	currentFD       *os.File
	currentFile     int
	currentFileSize int
	cache           *lru.Cache
	lock            *sync.RWMutex
	cacher          common.Cacher
}

type ReadInfo struct {
	fd     *os.File
	offset uint64
	size   uint64
}

func (ff *FlatFileManager) Truncate(lastIndex uint64) error {
	lastFile := lastIndex / ff.fileSizeLimit
	files, err := ioutil.ReadDir(ff.dataDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		name := filepath.Base(f.Name())
		i, err := strconv.ParseUint(name, 10, 64)
		if err == nil {
			if i < lastFile {
				err := os.Remove(path.Join(ff.dataDir, name))
				fmt.Println(err)
			}
		}
	}
	newff, err := NewFlatFile(ff.dataDir, ff.fileSizeLimit)
	if err != nil {
		return nil
	}
	*ff = *newff
	return nil
}

func (f *FlatFileManager) PasreFile(fileID int) (map[int]ReadInfo, error) {
	//TODO: cache
	fmt.Printf("[bmcachefile] hit/get %v/%v\n", totalHitFile, totalGetFile)
	f.lock.RLock()
	defer f.lock.RUnlock()

	if fileID < f.currentFile {
		v, ok := f.cache.Get(fileID)
		if ok {
			return v.(map[int]ReadInfo), nil
		}
	}

	p := path.Join(f.dataDir, strconv.Itoa(fileID))
	fd, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	readInfos := make(map[int]ReadInfo)
	offset := uint64(0)
	size := 0
	for {
		b := make([]byte, 8)
		n, _ := fd.ReadAt(b, int64(offset))
		var result uint64
		err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &result)
		if err != nil || n == 0 {
			break
		}

		readInf := ReadInfo{
			fd,
			offset + 8,
			result,
		}
		id := len(readInfos)
		readInfos[id] = readInf
		offset += 8
		offset += result
		size++
	}

	f.cache.Add(fileID, readInfos)

	return readInfos, nil
}

func (f FlatFileManager) Read(index uint64) ([]byte, error) {
	fileID := index / f.fileSizeLimit
	itemFileIndex := index % f.fileSizeLimit
	var readInfo map[int]ReadInfo
	var err error

	//find parse info in cache first
	v, ok := f.cache.Get(fileID)
	if ok {
		if itemFileIndex < len(v.(map[int]ReadInfo)) {
			readInfo = v.(map[int]ReadInfo)
		} else { //if find item index is greater than len of cache <- could be new item, parse again
			readInfo, err = f.PasreFile(fileID)
		}
	} else { //no cache, parse
		readInfo, err = f.PasreFile(fileID)
	}

	if err != nil {
		return nil, err
	}
	if (int(itemFileIndex) >= len(mapInfo)) || (!ok) {
		mapInfo, err = f.PasreFile(fileID)
		if err != nil {
			return nil, err
		}
	}

	if int(itemFileIndex) >= len(mapInfo) {
		return nil, errors.New(fmt.Sprintf("Cannot read item at index %v", index))
	}

	b := make([]byte, readInfo[itemFileIndex].size)
	readInfo[itemFileIndex].fd.ReadAt(b, int64(readInfo[itemFileIndex].offset))
	rawB, err := ioutil.ReadAll(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	return rawB, nil
}

func (f FlatFileManager) ReadRecently(index uint64) (chan []byte, chan int, func()) {
	c := make(chan []byte)
	e := make(chan uint64)
	closed := false

	var cancel = func() {
		if !closed {
			close(c)
			closed = true
		}

	}
	go func() {
		fromFile := index / uint64(f.fileSizeLimit)
		offset := index % uint64(f.fileSizeLimit)
		for i := int(fromFile); i >= 0; i-- {
			readInfo, err := f.PasreFile(int(i))
			if err != nil {
				e <- 1
				cancel()
			}
			if offset > uint64(len(readInfo)-1) {
				e <- 1
				cancel()
			}

			for j := int(offset); j >= 0; j-- {
				rawB := make([]byte, readInfo[int(j)].size)
				readInfo[int(j)].fd.ReadAt(rawB, int64(readInfo[int(j)].offset))

			LOOP:
				if !closed {
					select {
					case c <- rawB:
						continue
					default:
						time.Sleep(time.Millisecond * 10)
						goto LOOP
					}

				} else {
					return
				}

			}
		}
		cancel()
	}()

	return c, e, cancel
}

func (f *FlatFileManager) newNextFile() (*os.File, error) {
	i := uint64(0)
	if f.currentFD != nil {
		name := filepath.Base(f.currentFD.Name())
		i, _ = strconv.ParseUint(name, 10, 64)
		i++
	}
	path := path.Join(f.dataDir, strconv.FormatUint(i, 10))
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	f.currentFD = fd
	f.currentFile = i
	f.currentFileSize = 0
	f.sortedFolder = append(f.sortedFolder, i)
	f.folderMap[i] = true
	return fd, nil
}

//create new file if exceed size limit
func (f *FlatFileManager) update() error {
	size, err := f.checkFileSize()
	if err != nil {
		return err
	}
	if size >= f.fileSizeLimit {
		_, err := f.newNextFile()
		if err != nil {
			return err
		}
	} else {
		f.currentFileSize = size
	}
	return nil
}

func (f *FlatFileManager) checkFileSize() (uint64, error) {
	if f.currentFD == nil {
		return 0, errors.New("Not yet open file!")
	}
	offset := uint64(0)
	size := uint64(0)
	for {
		b := make([]byte, 8)
		n, _ := f.currentFD.ReadAt(b, int64(offset))
		var result uint64
		err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &result)
		if err != nil || n == 0 {
			break
		}

		offset += 8
		offset += result

		size++
	}
	return size, nil
}

func (f *FlatFileManager) Append(data []byte) (uint64, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	//append size-bytes into current FD, if max -> create new file, update currentFD
	var result = make([]byte, 8)
	binary.LittleEndian.PutUint64(result, uint64(len(data)))
	_, err := f.currentFD.Write(result)
	if err != nil {
		return 0, err
	}
	_, err = f.currentFD.Write(data)

	addedItemIndex := f.currentFile*f.fileSizeLimit + f.currentFileSize
	f.currentFileSize++
	if f.currentFileSize >= f.fileSizeLimit {
		f.newNextFile()
	}
	return addedItemIndex, err
}

func NewFlatFile(dir string, fileBound int) (*FlatFileManager, error) {
	cache, _ := lru.New(100)
	ff := &FlatFileManager{
		dataDir:       dir,
		fileSizeLimit: fileBound,
		folderMap:     make(map[uint64]bool),
		lock:          new(sync.RWMutex),
		cache:         cache,
	}

	//read all file has number  in folder -> into folderMap, sortedFolder
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	currentFile := uint64(0)
	needNewFile := true
	for _, f := range files {
		name := filepath.Base(f.Name())
		i, err := strconv.ParseUint(name, 10, 64)
		if err == nil {
			ff.folderMap[i] = true
			ff.sortedFolder = append(ff.sortedFolder, i)
			if currentFile < i {
				currentFile = i
				needNewFile = false
			}
		}
	}
	sort.Slice(ff.sortedFolder, func(i, j int) bool {
		if ff.sortedFolder[i] < ff.sortedFolder[j] {
			return true
		} else {
			return false
		}
	})
	//open last file currentFD, update currentFileSize, if max -> create new file, update currentFD
	if !needNewFile {
		path := path.Join(ff.dataDir, strconv.FormatUint(currentFile, 10))
		fd, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		ff.currentFD = fd
		ff.currentFile = currentFile
		//read file size, and create new file in need
		err = ff.update()
		if err != nil {
			return nil, err
		}
	} else {
		_, err := ff.newNextFile()
		if err != nil {
			return nil, err
		}
	}

	return ff, nil
}