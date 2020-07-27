# Pi Alive

A quick and dirty to broadcast your Pi to the world to make it easier for you to know if it's alive and what IP address it has.

Out of the box it publishes MQTT messages to both Hive and Mosquitto public MQTT brokers using your Pis MAC address in the topic.

## Note: Store your Pi MAC address to be able to subscribe to messages from it

## Setup

```sh
cd /tmp
curl -L https://github.com/ejdoh1/pialive/raw/master/pialive --output pialive
chmod +x pialive
cat <<EOT >> /etc/init.d/pialive.sh
export PIALIVE_BASE64_ENCODE=false
export PIALIVE_COMMAND="ifconfig && iwgetid"
./tmp/pialive
EOT
reboot

```
