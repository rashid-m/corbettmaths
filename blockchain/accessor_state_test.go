package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"os"
	"reflect"
	"testing"
)

func TestStoreStateObjectToFlatFile(t *testing.T) {

	os.RemoveAll("./tmp")
	defer os.RemoveAll("./tmp")
	flatFile, err := flatfile.NewFlatFile("./tmp", 5000)
	if err != nil {
		t.Fatal(err)
	}
	storeObject := [][]byte{
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 90, 38, 90, 72, 2, 80, 47, 255, 11, 61, 102, 58, 136, 196, 243, 248, 175, 183, 92, 230, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 112, 72, 50, 100, 119, 111, 101, 56, 116, 117, 50, 81, 102, 69, 77, 113, 81, 122, 110, 78, 53, 115, 83, 57, 109, 72, 83, 116, 107, 72, 98, 119, 84, 66, 69, 72, 53, 55, 82, 99, 55, 56, 51, 118, 70, 107, 90, 68, 67, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 50, 246, 115, 201, 67, 201, 75, 94, 181, 107, 158, 151, 48, 100, 221, 122, 199, 219, 137, 62, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 49, 100, 120, 115, 75, 50, 74, 56, 89, 85, 75, 89, 105, 65, 102, 88, 98, 101, 121, 89, 101, 50, 100, 90, 99, 102, 112, 75, 71, 57, 49, 86, 81, 122, 101, 98, 78, 71, 122, 52, 119, 97, 118, 102, 107, 66, 118, 104, 52, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 93, 26, 55, 122, 166, 192, 26, 217, 252, 62, 189, 252, 195, 219, 202, 168, 69, 128, 219, 232, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 51, 52, 56, 49, 57, 49, 56, 51, 50, 54, 48, 48, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 86, 56, 89, 66, 68, 120, 77, 97, 66, 84, 55, 121, 84, 113, 67, 77, 65, 80, 114, 101, 67, 121, 113, 120, 83, 76, 89, 99, 52, 101, 50, 72, 97, 99, 57, 77, 83, 110, 72, 110, 80, 56, 106, 114, 120, 101, 85, 85, 53, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 89, 46, 135, 226, 89, 30, 225, 74, 181, 202, 3, 53, 82, 120, 240, 5, 132, 17, 15, 197, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 103, 111, 49, 52, 118, 66, 115, 72, 116, 115, 103, 51, 71, 89, 83, 77, 84, 111, 100, 120, 54, 114, 75, 113, 68, 51, 85, 99, 116, 119, 109, 72, 88, 80, 51, 103, 88, 70, 103, 88, 90, 107, 69, 97, 55, 76, 105, 110, 76, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 225, 121, 133, 142, 175, 49, 39, 60, 67, 130, 220, 54, 191, 241, 139, 82, 41, 156, 89, 118, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 74, 78, 57, 75, 82, 103, 105, 69, 122, 70, 78, 65, 49, 50, 90, 69, 69, 69, 113, 85, 52, 69, 51, 115, 77, 74, 102, 90, 119, 72, 84, 53, 99, 101, 110, 90, 57, 83, 80, 53, 115, 69, 103, 49, 111, 104, 106, 74, 119, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 164, 33, 240, 49, 205, 159, 29, 179, 118, 32, 59, 120, 208, 212, 70, 181, 34, 96, 164, 210, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 107, 65, 49, 101, 89, 51, 83, 103, 104, 55, 68, 117, 107, 75, 99, 113, 57, 81, 111, 119, 66, 105, 78, 52, 71, 98, 113, 68, 105, 99, 67, 122, 77, 100, 119, 52, 85, 54, 90, 49, 121, 81, 56, 51, 85, 88, 55, 75, 102, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 130, 149, 32, 12, 126, 5, 199, 99, 124, 55, 247, 40, 29, 206, 175, 241, 227, 149, 254, 243, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 115, 84, 109, 81, 120, 52, 122, 89, 82, 68, 119, 70, 49, 100, 74, 103, 110, 65, 86, 109, 77, 72, 111, 100, 121, 75, 57, 117, 118, 69, 89, 71, 50, 70, 110, 67, 121, 65, 105, 83, 66, 75, 122, 72, 115, 112, 53, 105, 82, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 255, 31, 100, 61, 100, 30, 220, 163, 205, 208, 98, 219, 159, 161, 124, 62, 231, 148, 69, 168, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 86, 85, 81, 105, 80, 122, 118, 114, 75, 52, 71, 120, 109, 82, 86, 69, 57, 53, 121, 104, 89, 49, 117, 121, 118, 77, 87, 49, 101, 102, 122, 52, 71, 104, 82, 98, 110, 68, 74, 89, 72, 89, 80, 113, 117, 88, 78, 68, 107, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 91, 209, 124, 67, 116, 94, 43, 43, 134, 54, 61, 145, 42, 154, 161, 198, 246, 111, 169, 214, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 67, 80, 49, 69, 85, 105, 114, 81, 122, 66, 49, 109, 75, 66, 86, 82, 82, 103, 51, 69, 69, 107, 86, 88, 77, 110, 106, 119, 111, 74, 88, 89, 117, 54, 67, 81, 97, 75, 67, 68, 78, 115, 104, 103, 115, 84, 107, 119, 90, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 91, 63, 149, 37, 235, 54, 76, 210, 242, 217, 132, 71, 237, 201, 113, 235, 152, 165, 89, 164, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 116, 100, 85, 114, 107, 52, 54, 55, 115, 100, 54, 88, 111, 88, 75, 86, 109, 107, 115, 56, 89, 114, 78, 52, 49, 78, 85, 80, 72, 89, 121, 89, 68, 78, 116, 109, 72, 119, 55, 55, 118, 83, 113, 121, 99, 74, 82, 71, 74, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 184, 89, 90, 114, 197, 179, 24, 84, 124, 195, 231, 46, 217, 27, 116, 71, 102, 91, 113, 70, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 68, 66, 121, 86, 74, 106, 122, 88, 70, 72, 66, 101, 101, 51, 89, 55, 118, 102, 81, 107, 118, 74, 85, 71, 84, 101, 106, 99, 78, 53, 54, 105, 97, 116, 100, 110, 55, 72, 81, 74, 112, 77, 69, 84, 104, 121, 103, 56, 83, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 224, 109, 173, 49, 85, 94, 61, 239, 82, 114, 229, 38, 243, 187, 103, 197, 59, 181, 71, 117, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 81, 122, 104, 121, 87, 83, 104, 109, 119, 106, 70, 69, 97, 115, 107, 109, 68, 65, 106, 122, 113, 104, 84, 84, 56, 65, 74, 116, 70, 70, 89, 107, 101, 115, 68, 88, 114, 86, 75, 115, 97, 121, 82, 78, 87, 98, 113, 110, 72, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 33, 74, 124, 148, 134, 187, 240, 12, 113, 250, 97, 243, 148, 22, 230, 116, 101, 187, 10, 125, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 114, 103, 111, 116, 66, 56, 98, 99, 121, 122, 52, 78, 66, 110, 74, 65, 69, 116, 81, 105, 100, 86, 72, 82, 112, 83, 82, 52, 72, 87, 72, 106, 53, 68, 122, 54, 85, 75, 97, 68, 98, 103, 120, 65, 112, 77, 49, 104, 66, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 124, 6, 102, 245, 133, 29, 88, 177, 146, 134, 169, 211, 83, 83, 190, 172, 108, 31, 217, 227, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 102, 69, 85, 99, 86, 51, 112, 114, 67, 87, 67, 68, 78, 78, 51, 110, 52, 88, 105, 88, 117, 86, 49, 67, 122, 75, 122, 116, 87, 115, 84, 70, 99, 76, 106, 65, 55, 49, 100, 106, 89, 84, 82, 65, 116, 76, 122, 97, 109, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 149, 252, 107, 184, 165, 181, 56, 56, 29, 187, 68, 57, 93, 138, 165, 104, 14, 21, 98, 99, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 51, 117, 118, 57, 68, 49, 80, 85, 101, 55, 76, 112, 119, 122, 101, 85, 57, 51, 110, 53, 100, 72, 104, 52, 89, 116, 119, 88, 80, 117, 51, 82, 52, 54, 49, 87, 77, 116, 86, 112, 119, 107, 76, 72, 109, 118, 81, 70, 74, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 207, 204, 23, 255, 156, 222, 29, 70, 15, 61, 229, 41, 100, 35, 57, 16, 113, 120, 85, 203, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 82, 82, 77, 71, 65, 69, 88, 100, 116, 68, 55, 57, 54, 115, 98, 109, 121, 72, 119, 78, 118, 57, 111, 105, 80, 80, 102, 71, 89, 97, 49, 114, 54, 75, 115, 115, 77, 113, 88, 121, 105, 85, 77, 102, 84, 54, 114, 56, 54, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 65, 39, 40, 94, 29, 69, 188, 86, 192, 59, 218, 166, 122, 225, 182, 136, 171, 147, 243, 48, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 80, 49, 53, 117, 102, 78, 82, 90, 76, 72, 55, 54, 104, 51, 118, 90, 122, 80, 114, 71, 51, 57, 69, 68, 116, 102, 86, 122, 106, 66, 90, 87, 55, 68, 68, 119, 83, 85, 87, 103, 75, 78, 55, 75, 114, 105, 65, 78, 104, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 102, 20, 17, 149, 33, 181, 198, 111, 8, 205, 203, 68, 20, 35, 44, 179, 228, 112, 216, 91, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 97, 122, 81, 120, 74, 109, 70, 103, 77, 120, 76, 105, 97, 90, 103, 109, 78, 103, 117, 112, 71, 110, 118, 70, 86, 53, 98, 119, 82, 78, 118, 77, 109, 69, 53, 89, 97, 82, 116, 87, 112, 84, 80, 50, 65, 70, 57, 68, 82, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 5, 151, 212, 100, 24, 31, 33, 28, 132, 198, 169, 150, 189, 204, 41, 250, 215, 191, 166, 209, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 105, 76, 106, 76, 85, 101, 86, 118, 117, 88, 50, 90, 84, 77, 97, 89, 85, 57, 113, 88, 112, 103, 111, 117, 55, 86, 67, 90, 116, 68, 81, 67, 83, 113, 118, 71, 57, 97, 83, 82, 68, 84, 71, 117, 65, 99, 55, 49, 76, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 221, 54, 151, 249, 230, 233, 19, 110, 133, 67, 240, 68, 56, 155, 2, 10, 198, 121, 204, 181, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 90, 84, 97, 118, 52, 72, 52, 116, 115, 56, 81, 86, 117, 52, 90, 66, 104, 78, 97, 78, 90, 50, 107, 72, 52, 56, 76, 102, 107, 71, 109, 71, 97, 55, 89, 78, 66, 86, 50, 70, 107, 111, 87, 84, 53, 71, 99, 84, 112, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 161, 117, 104, 3, 55, 48, 86, 2, 216, 168, 78, 192, 74, 172, 89, 42, 244, 28, 189, 177, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 120, 69, 106, 74, 84, 83, 55, 57, 101, 119, 88, 55, 114, 86, 82, 89, 52, 52, 57, 81, 84, 69, 53, 84, 50, 66, 57, 98, 117, 121, 99, 49, 49, 98, 86, 114, 49, 78, 55, 71, 69, 122, 80, 66, 89, 121, 85, 100, 56, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 12, 207, 232, 4, 200, 49, 168, 75, 33, 254, 28, 30, 163, 246, 143, 228, 213, 42, 220, 26, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 71, 107, 70, 83, 75, 111, 90, 70, 76, 100, 78, 117, 57, 80, 107, 116, 115, 107, 120, 77, 90, 51, 51, 89, 103, 99, 100, 111, 53, 105, 70, 57, 50, 72, 113, 87, 72, 87, 75, 84, 106, 70, 116, 114, 105, 121, 68, 86, 49, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 82, 249, 160, 67, 85, 113, 170, 196, 246, 84, 94, 207, 78, 99, 166, 247, 220, 57, 80, 179, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 72, 69, 72, 105, 117, 114, 121, 55, 78, 77, 98, 90, 109, 120, 89, 107, 110, 67, 55, 118, 113, 104, 114, 99, 102, 68, 99, 98, 121, 103, 56, 82, 78, 50, 52, 74, 115, 120, 50, 86, 54, 111, 97, 49, 117, 82, 103, 119, 88, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 115, 34, 68, 47, 164, 133, 22, 80, 202, 25, 174, 32, 88, 39, 136, 253, 1, 151, 159, 120, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 86, 86, 81, 82, 71, 90, 70, 104, 77, 121, 112, 57, 76, 82, 83, 49, 121, 106, 121, 88, 115, 87, 54, 117, 103, 77, 113, 116, 72, 97, 89, 121, 85, 89, 77, 76, 117, 116, 82, 104, 121, 54, 65, 83, 70, 55, 65, 87, 102, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 149, 175, 35, 43, 30, 106, 40, 166, 233, 207, 217, 76, 16, 116, 246, 138, 71, 254, 137, 66, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 120, 98, 54, 57, 65, 57, 99, 69, 89, 117, 120, 52, 88, 67, 102, 101, 80, 49, 52, 83, 107, 51, 111, 57, 114, 103, 55, 83, 122, 80, 106, 67, 85, 115, 115, 68, 119, 69, 67, 72, 102, 117, 67, 57, 68, 122, 104, 100, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 130, 162, 114, 76, 200, 59, 239, 1, 181, 68, 30, 98, 75, 246, 243, 182, 141, 54, 19, 92, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 53, 111, 113, 83, 112, 118, 86, 121, 53, 74, 80, 104, 55, 105, 77, 53, 72, 101, 100, 88, 101, 107, 116, 89, 99, 121, 84, 100, 86, 49, 97, 116, 52, 103, 109, 103, 102, 82, 75, 116, 51, 109, 75, 103, 121, 72, 118, 118, 52, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 166, 76, 206, 182, 71, 132, 75, 183, 93, 218, 26, 62, 50, 65, 115, 157, 161, 54, 183, 20, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 65, 50, 72, 84, 84, 75, 105, 84, 107, 49, 98, 103, 65, 114, 66, 90, 121, 111, 114, 69, 65, 105, 78, 86, 101, 54, 98, 81, 97, 76, 114, 102, 87, 49, 82, 80, 88, 53, 99, 85, 90, 78, 102, 86, 75, 98, 53, 100, 53, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 50, 29, 190, 254, 74, 36, 85, 252, 190, 86, 75, 95, 124, 73, 92, 183, 218, 21, 32, 224, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 70, 122, 111, 52, 70, 78, 77, 82, 55, 114, 121, 110, 114, 83, 99, 50, 110, 104, 106, 78, 54, 49, 99, 78, 100, 107, 70, 121, 98, 76, 97, 69, 106, 84, 100, 90, 52, 84, 76, 111, 99, 77, 76, 117, 116, 69, 98, 86, 122, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 14, 213, 65, 183, 102, 204, 22, 67, 231, 237, 109, 71, 170, 99, 102, 155, 144, 214, 174, 140, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 49, 53, 48, 50, 49, 51, 55, 52, 53, 57, 54, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 88, 87, 52, 114, 67, 106, 105, 85, 87, 119, 105, 87, 80, 78, 78, 85, 84, 112, 83, 66, 66, 97, 85, 70, 86, 66, 105, 106, 120, 117, 100, 77, 113, 86, 76, 57, 118, 112, 70, 97, 52, 80, 76, 71, 52, 101, 51, 81, 117, 34, 125},
		{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 233, 100, 107, 8, 51, 71, 173, 112, 35, 64, 246, 32, 198, 249, 241, 100, 80, 34, 148, 24, 56, 20, 184, 91, 149, 183, 158, 188, 208, 237, 163, 123, 34, 82, 101, 119, 97, 114, 100, 34, 58, 123, 34, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 52, 34, 58, 56, 57, 53, 51, 53, 48, 52, 50, 54, 54, 56, 125, 44, 34, 73, 110, 99, 111, 103, 110, 105, 116, 111, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 49, 50, 57, 53, 67, 50, 89, 52, 78, 69, 118, 71, 54, 87, 122, 111, 69, 116, 67, 53, 88, 74, 99, 105, 120, 113, 103, 111, 53, 81, 109, 52, 66, 87, 89, 54, 81, 118, 57, 105, 56, 68, 50, 86, 49, 81, 80, 56, 88, 121, 88, 34, 125},
	}

	storeIndexes := []int{}

	for _, v := range storeObject {
		newIndex, err := flatFile.Append(v)
		if err != nil {
			t.Fatal(err)
		}
		storeIndexes = append(storeIndexes, newIndex)
	}

	getObject := [][]byte{}

	for _, v := range storeIndexes {
		res, err := flatFile.Read(v)
		if err != nil {
			t.Fatal(err)
		}
		getObject = append(getObject, res)
	}

	t.Log(reflect.DeepEqual(getObject, storeObject))

}