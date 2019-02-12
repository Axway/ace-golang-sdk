rm -rf /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk
mkdir /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk

cp -r ./config    /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk/config 
cp -r ./linker    /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk/linker 
cp -r ./messaging /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk/messaging 
cp -r ./rpc       /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk/rpc
cp -r ./util      /home/swav/dev/go/src/git.ecd.axway.int/choreo/stomp/vendor/github.com/Axway/ace-golang-sdk/util
echo "done copying", `date`
