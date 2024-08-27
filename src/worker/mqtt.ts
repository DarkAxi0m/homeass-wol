import mqtt, { MqttClient } from "mqtt";

const host: string = "10.1.1.36";
const port: string = "1883";
const prefix: string = "homeassistant";

const clientId: string = `mqtt_${Math.random().toString(16).slice(3)}`;

const connectUrl: string = `mqtt://${host}:${port}`;

console.log(connectUrl);

const client: MqttClient = mqtt.connect(connectUrl, {
  clientId,
  clean: true,
  connectTimeout: 4000,
  reconnectPeriod: 1000,
});

class DeviceType {
  manufacturer: string;
  model: string;
  name: string;
  identifiers: string[];
  constructor(name: string, identifiers: string[]) {
    this.manufacturer = "SiRMonkeys";
    this.name = name;
    this.model = "MQTTWOL";
    this.identifiers = identifiers;
  }
}

type DeviceTopic = {
  config_topic: string;
  state_topic: string;
  command_topic?: string;
};

class MQTTDevice {
  objectid: string;
  switchId: string = "PowerSW";
  onOffId: string = "OnOff";
  device: DeviceType;

  topicSwitch: DeviceTopic;
  topicOnOff: DeviceTopic;

  constructor(name: string) {
    this.objectid = name.replace(" ", "_").toLowerCase();
    this.device = new DeviceType(name, [this.objectid]);

    this.topicSwitch = {
      config_topic: `${prefix}/switch/${this.objectid}/${this.switchId}/config`,
      state_topic: `${prefix}/switch/${this.objectid}/${this.switchId}/state`,
      command_topic: `${prefix}/binary_sensor/${this.objectid}/${this.switchId}/set`,
    };
    this.topicOnOff = {
      config_topic: `${prefix}/binary_sensor/${this.objectid}/${this.onOffId}/config`,
      state_topic: `${prefix}/binary_sensor/${this.objectid}/${this.onOffId}/state`,
    };
  }

  publish(client: MqttClient) {
    client.publish(
      this.topicSwitch.config_topic,
      JSON.stringify({
        name: "Power Switch",
        command_topic: this.topicSwitch.command_topic,
        state_topic: this.topicSwitch.state_topic,
        unique_id: `${this.objectid}_${this.switchId}`,
        device: this.device,
      }),
      { qos: 0, retain: false },
    );

    client.publish(
      this.topicOnOff.config_topic,
      JSON.stringify({
        name: "Running",
        device_class: "running",
        state_topic: this.topicOnOff.state_topic,
        unique_id: `${this.objectid}_${this.onOffId}`,
        device: this.device,
      }),
      { qos: 0, retain: false },
    );

    client.subscribe([this.topicSwitch.command_topic]);

    client.publish(
      this.topicOnOff.state_topic,

      "OFF",
      { qos: 0, retain: false },
    );
    client.publish(
      this.topicSwitch.state_topic,

      "OFF",
      { qos: 0, retain: false },
    );
  }

  onMessage(topic: string, payload: Buffer):boolean {
    if (topic == this.topicSwitch.command_topic) {
      setTimeout(() => {
        client.publish(
          this.topicSwitch.state_topic,

          payload.toString(),
          { qos: 0, retain: false },
        );
      }, 2000);

      setTimeout(() => {
        client.publish(
          this.topicOnOff.state_topic,

          payload.toString(),
          { qos: 0, retain: false },
        );
      }, 1000);
      
      return true;
    }
    return false;
  }
}

const nasdevice = new MQTTDevice("Ture NAS");
const testdevice = new MQTTDevice("Test Device");

client.on("connect", () => {
  console.log("Connected");

  client.subscribe(["homeassistant/status"]); //need to react to this

  nasdevice.publish(client);
  testdevice.publish(client);
});

client.on("message", (topic: string, payload: Buffer) => {
  console.log("----");
  console.log("📲", topic);
  console.log(payload.toString());

  nasdevice.onMessage(topic, payload);
  testdevice.onMessage(topic, payload);
});

/* */
