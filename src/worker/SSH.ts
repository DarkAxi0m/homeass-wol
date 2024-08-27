import { exec } from "child_process";

interface SSHConfig {
  username: string;
  host: string;
  privateKeyPath: string;
  password?: string;
}

export function shutdownServer(config: SSHConfig): Promise<void> {
  return new Promise((resolve, reject) => {
    const sudoCommand = config.password
      ? `echo ${config.password} | sudo -S shutdown -h now`
      : `sudo shutdown -h now`;
    console.log("Shutdown Server,", config);

    const command = `ssh -i ${config.privateKeyPath} ${config.username}@${config.host} '${sudoCommand}'`;

    exec(command, (error, stdout, stderr) => {
      if (error) {
        reject(`Error: ${stderr}`);
      } else {
        resolve();
      }
    });
  });
}
