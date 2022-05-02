#!/bin/bash

# ============================================================================================
# 1. Interactive mode: just run following command then follow the steps.
#       sudo ./{this script}
#	2. Preconfig mode: change the below configs, then run:
#    		sudo ./{this script} -y
#	3. To uninstall, run:
#		    sudo ./{this script} -u
# 4. If you want to add more node after running this script:
#      4.1. Open /home/incognito/validator_keys, append more keys to it.
#          (separate by commas, no spaces)
#      4.2. Delete /home/incognito/inc_node_latest_tag
#      4.3. Start IncognitoUpdater service again:
#          sudo systemctl start IncognitoUpdater.service
#        or just run the run_node.sh script:
#          sudo /home/incognito/run_node.sh
# ============================================================================================

# ============================= CHANGE CONFIG HERE ===========================================
VALIDATOR_K=("validator_key_1,validator_key_2,validator_key_3"
  "! Input validator keys here, multiple validator keys must be separated by commas (no spaces):\n\t> ")
GETH_NAME=("https://mainnet.infura.io/v3/xxxyyy"
  "! Infura link. Example: 'https://mainnet.infura.io/v3/xxxyyy'.
   (follow step 3 on this thread to setup infura https://we.incognito.org/t/194):\n\t> ")
PORT_RPC=("8334"
  "! RPC port, should be left as default (8334),
   The first node uses this port, the next one uses port+1 and so on: ")
PORT_NODE=("9433"
  "! Node port, should be left as default (9334),
   The first node uses this port, the next one uses port+1 and so on: ")

# New parameters since privacy v2
NUM_INDEXER_WORKERS=(100
  "! Number of coin indexer worker, default = 100. To disable this, set it to 0: ")
INDEXER_ACCESS_TOKEN=("edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb"
  "! Indexer access token, can be generated by running: $ echo 'bla bla bla' | sha256sum
    (default: edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb):\n\t> ")
# =================================== END CONFIG ============================================

# ============================================================================================
# Do not edit lines below unless you know what you're doing
BOOTNODE="mainnet-bootnode.incognito.org:9330"  # this should be left as default
FULLNODE=""  # set to 1 to run as a full node, empty to run as normal node
SERVICE="/etc/systemd/system/IncognitoUpdater.service"
TIMER="/etc/systemd/system/IncognitoUpdater.timer"
USER_NAME="incognito"
INC_HOME="/home/$USER_NAME"
DATA_DIR="$INC_HOME/node_data"
TMP="$INC_HOME/inc_node_latest_tag"
KEY_FILE="$INC_HOME/validator_keys"
SCRIPT="$INC_HOME/run_node.sh"
DBMODE="archive"
FFSTORAGE="false"

# check super user
if [ $(whoami) != root ]; then
cat << EOF
	!!! Please run with sudo or su, otherwise it won't work
	!!! Script now exits
EOF
	exit 1
fi

function uninstall {
cat << EOF
!!!===============================================!!!
###                    WARNING                    ###
!!!===============================================!!!
!   Uninstalling and cleanup !
!   This action will remove:
!      - The systemd service and timer: $SERVICE $TIMER
!      - Docker containers and images
!      - User: $USER_NAME
!      - $INC_HOME (including all node's data, logs and everything else inside the folder)
!      - Run script: $SCRIPT
!!! Do you really want to do this? (N/y)
EOF
	read consent
	if [ -z $consent ] || [[ ${consent,,} = "n" ]] || [[ ${consent,,} != "y" ]] ; then
		echo "!!! Good choice !!!"
		exit 1
	else
cat << EOF
#####################################################
#          Too bad! So sad! See you again!          #
#####################################################
EOF
    echo " # Remove update service"
    systemctl stop $(basename $SERVICE) 2> /dev/null
    systemctl stop $(basename $TIMER) 2> /dev/null
    systemctl disable $(basename $SERVICE)
    systemctl disable $(basename $TIMER)
    systemctl daemon-reload
    echo " # Stop and remove docker images + containers"
    docker container stop $(docker container ls -aqf name=inc_mainnet)
    docker container rm $(docker container ls -aqf name=inc_mainnet)
    docker image rm -f $(docker images -q incognitochain/incognito-mainnet)
    echo " # Removing user"
    deluser $USER_NAME
    echo " # Removing $INC_HOME (including all node's data, logs and everything else inside the folder)"
    rm -Rf $INC_HOME $SERVICE $TIMER
    exit 1
  fi
}

# Parsing args
interactive_mode=1
while getopts "uyh" option; do
   case "$option" in
      "h")
cat << EOF
   $(basename $0) -u|y|h
      -h: print this help message then exit.
      -y: Run setup with all default settings in the script, non-interactive mode
      -u: Uninstall auto update service

   Example:
    To uninstall the auto update service
      $(basename $0) -u

   or
    To install the auto update service with non-interactive mode
      $(basename $0) -y

EOF
         exit
         ;;
      "u")
         uninstall
         ;;
      "y")
         interactive_mode=""
         ;;
      ?)
         exit
         ;;
   esac
done

echo "Checking for / Installing Docker and jq (JSON processor)"
apt install docker.io jq -y
systemctl start docker.service
echo "Finished. Launching setup script..."
sleep 3
clear
cat << 'EOF'
      ██╗███╗░░██╗░█████╗░░█████╗░░██████╗░███╗░░██╗██╗████████╗░█████╗░░░░░█████╗░██████╗░░██████╗░
      ██║████╗░██║██╔══██╗██╔══██╗██╔════╝░████╗░██║██║╚══██╔══╝██╔══██╗░░░██╔══██╗██╔══██╗██╔════╝░
      ██║██╔██╗██║██║░░╚═╝██║░░██║██║░░██╗░██╔██╗██║██║░░░██║░░░██║░░██║░░░██║░░██║██████╔╝██║░░██╗░
      ██║██║╚████║██║░░██╗██║░░██║██║░░╚██╗██║╚████║██║░░░██║░░░██║░░██║░░░██║░░██║██╔══██╗██║░░╚██╗
      ██║██║░╚███║╚█████╔╝╚█████╔╝╚██████╔╝██║░╚███║██║░░░██║░░░╚█████╔╝██╗╚█████╔╝██║░░██║╚██████╔╝
      ╚═╝╚═╝░░╚══╝░╚════╝░░╚════╝░░╚═════╝░╚═╝░░╚══╝╚═╝░░░╚═╝░░░░╚════╝░╚═╝░╚════╝░╚═╝░░╚═╝░╚═════╝░
            __    __                  __                   ______                       __              __
           /  \  /  |                /  |                 /      \                     /  |            /  |
 __     __ $$  \ $$ |  ______    ____$$ |  ______        /$$$$$$  |  _______   ______  $$/   ______   _$$ |_
/  \   /  |$$$  \$$ | /      \  /    $$ | /      \       $$ \__$$/  /       | /      \ /  | /      \ / $$   |
$$  \ /$$/ $$$$  $$ |/$$$$$$  |/$$$$$$$ |/$$$$$$  |      $$      \ /$$$$$$$/ /$$$$$$  |$$ |/$$$$$$  |$$$$$$/
 $$  /$$/  $$ $$ $$ |$$ |  $$ |$$ |  $$ |$$    $$ |       $$$$$$  |$$ |      $$ |  $$/ $$ |$$ |  $$ |  $$ | __
  $$ $$/   $$ |$$$$ |$$ \__$$ |$$ \__$$ |$$$$$$$$/       /  \__$$ |$$ \_____ $$ |      $$ |$$ |__$$ |  $$ |/  |
   $$$/    $$ | $$$ |$$    $$/ $$    $$ |$$       |      $$    $$/ $$       |$$ |      $$ |$$    $$/   $$  $$/
    $/     $$/   $$/  $$$$$$/   $$$$$$$/  $$$$$$$/        $$$$$$/   $$$$$$$/ $$/       $$/ $$$$$$$/     $$$$/
                                                                                           $$ |
                                                                                           $$ |
                                                                                           $$/
EOF
if [[ $interactive_mode ]]; then # interactive mode, taking user input
cat << EOF
=======================================================
     Start setup with interactive mode
=======================================================

!! To select default values, just hit ENTER !!

EOF
  printf "${VALIDATOR_K[1]}"
  read VALIDATOR_K[0]

  printf "${GETH_NAME[1]}"
  read GETH_NAME[0]

  printf "${PORT_RPC[1]}"
  read input
  if [[ ! -z $input ]]; then PORT_RPC[0]=$input; fi

  printf "${PORT_NODE[1]}"
  read input
  if [[ ! -z $input ]]; then PORT_NODE[0]=$input; fi

  printf "${NUM_INDEXER_WORKERS[1]}"
  read input
  if [[ ! -z $input ]]; then NUM_INDEXER_WORKERS[0]=$input; fi

  printf "${INDEXER_ACCESS_TOKEN[1]}"
  read input
  if [[ ! -z $input ]]; then INDEXER_ACCESS_TOKEN[0]=$input; fi

else
cat << EOF
=======================================================
     Start setup with non-interactive mode
=======================================================
EOF
fi

cat << EOF
Configurations:
      Validator keys: ${VALIDATOR_K[0]}
      Infura: ${GETH_NAME[0]}
      RPC port: ${PORT_RPC[0]}
      Node port: ${PORT_NODE[0]}
      Number of indexer worker: ${NUM_INDEXER_WORKERS[0]}
      Coin indexer access token: ${INDEXER_ACCESS_TOKEN[0]}
EOF

systemctl stop $(basename $SERVICE) 2> /dev/null
systemctl stop $(basename $TIMER) 2> /dev/null
echo " # Creating $USER_NAME user to run node"
useradd $USER_NAME
usermod -aG docker $USER_NAME || echo
mkdir -p $INC_HOME
chown -R $USER_NAME:$USER_NAME $INC_HOME
rm -f $TMP && touch $TMP
chown $USER_NAME:$USER_NAME $TMP
echo ${VALIDATOR_K[0]} > $KEY_FILE

echo " # Creating systemd service to check for new release"
cat << EOF > $SERVICE
[Unit]
Description = IncognitoChain Node updater
After = network.target network-online.target
Wants = network-online.target

[Service]
Type = oneshot
User = $USER_NAME
ExecStart = $SCRIPT
StandardOutput = syslog
StandardError = syslog
SyslogIdentifier = IncNodeUpdt

[Install]
WantedBy = multi-user.target
EOF

echo " # Creating timer to preodically run the update checker"
cat << EOF > $TIMER
[Unit]
Description=Run IncognitoUpdater hourly

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
EOF

echo " # Creating run node script"
cat << EOF > $SCRIPT
#!/bin/bash
TMP=$TMP
key_file=$KEY_FILE
run()
{
  bootnode=$BOOTNODE
  dbmode=$DBMODE
  ffstorage=$FFSTORAGE
  data_dir=$DATA_DIR
  rpc_port=${PORT_RPC[0]}
  node_port=${PORT_NODE[0]}
  geth_name=${GETH_NAME[0]}
  geth_port=""
  geth_proto=""
  fullnode=$FULLNODE
  coin_index_access_token=${INDEXER_ACCESS_TOKEN[0]}
  num_index_worker=${NUM_INDEXER_WORKERS[0]}
EOF

cat << 'EOF' >> $SCRIPT
  validator_key=$(cat $key_file)
  validator_key=(${validator_key//,/ })
  latest_tag=$1
  current_tag=$2
  backup_log=0
  container_name=inc_mainnet
  count=0

  if [ -z "$node_port" ]; then
    node_port="9433";
  fi
  if [ -z "$rpc_port" ]; then
    rpc_port="8334";
  fi

  if [ "$current_tag" != "" ]; then
	  echo "Found new docker tag, remove old one"
    docker image rm -f incognitochain/incognito-mainnet:${current_tag}
  fi

  echo "Pulling new tag: ${latest_tag}"
  docker pull incognitochain/incognito-mainnet:${latest_tag}
  echo "Create new docker network"
  docker network create --driver bridge inc_net || true
  for key in ${validator_key[@]}; do
    echo "Remove old container"
    docker container stop ${container_name}_${count}
    docker container rm ${container_name}_${count}

    echo "Start the incognito mainnet docker container"
    set -x
    docker run --restart=always --net inc_net \
      -p $node_port:$node_port -p $rpc_port:$rpc_port \
      -e DBMODE=$dbmode -e FFStorage=$ffstorage \
      -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode \
      -e GETH_NAME=$geth_name -e GETH_PROTOCOL=$geth_proto -e GETH_PORT=$geth_port \
      -e FULLNODE=$fullnode -e MININGKEY=${key} -e TESTNET=false -e LIMIT_FEE=1 \
      -e INDEXER_ACCESS_TOKEN=$coin_index_access_token -e NUM_INDEXER_WORKERS=$num_index_worker \
      -v ${data_dir}_${count}:/data -d --name ${container_name}_${count} incognitochain/incognito-mainnet:${latest_tag}
    set +x
    ((node_port++))
    ((rpc_port++))
    ((count++))
  done
}

current_latest_tag=$(cat $TMP)
echo "Getting Incognito docker tags"
tags=$(curl -s -X GET https://hub.docker.com/v1/repositories/incognitochain/incognito-mainnet/tags | jq ".[].name")
tags=${tags//\"/}
sorted_tags=($(echo ${tags[*]}| tr " " "\n" | sort -rn))
latest_tag=${sorted_tags[0]}
echo "Latest tag is ${latest_tag}"

if [ "$current_latest_tag" != "$latest_tag" ]; then
	echo "Found newer tag , run it!"
	run $latest_tag $current_latest_tag
	current_latest_tag=$latest_tag
	echo $current_latest_tag > $TMP
else
    echo "Runing latest tag already, no need to update"
fi
EOF

chmod +x $SCRIPT

echo " # Enabling service"
systemctl daemon-reload
systemctl enable $(basename $SERVICE)
systemctl enable $(basename $TIMER)


echo " # Starting service. Please wait..."
systemctl start $(basename $SERVICE)
systemctl start $(basename $TIMER)
cat << EOF
 # DONE.
 To check the installing and starting progress or the running service:
    $ journalctl | grep Inc
    or
    $ journalctl -t IncNodeUpdt
EOF
