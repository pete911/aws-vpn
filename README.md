# aws-vpn

Currently only OpenVPN is supported. Project creates t3.micro EC2 instance with amazon linux 2023 (x86_64 architecture).
[User data](internal/vpn/ovpn/tmpl.go) is executed and VPN is configured. All secrets (certificates and keys) are stored
in secrets manager. Security groups are configured to allow only users (whoever runs the cli) IP. This can be overridden
with `--inbound-cidr` flag if needed, when creating VPN.

```
                 ┌──────────┐                            
                 │          │                            
        ┌────────► Secrets  │                            
        │        │ Manager  │                            
        │        └──────────┘                            
   ┌────┴────┐   ┌──────────┐                            
   │         │   │          │                            
   │   EC2   ┼───► Security │                            
   │  (VPN)  │   │ Group    │                            
   └────┬────┘   └──────────┘                            
        │        ┌──────────┐  ┌──────────┐              
        │        │          │  │          │              
        └────────► Instance ┼──► IAM role │              
                 │ Profile  │  │          │              
                 └──────────┘  └──────────┘              
```

## pre-requisites

AWS account has to have default VPC with access to internet. To keep this project simple (and cost-effective), it does
not create VPC (IGW and/or NAT gateways), but re-uses existing default VPC.

## usage

If you have more than one AWS account, you can use different profile by prefixing commands with `AWS_PROFILE=<profiel> <cmd>`
- create VPN `aws-vpn create <name>` (name has to be unique per region)
  - you can either select region via `--region` flag, otherwise prompt with select is presented
- get client config `aws-vpn config` (either select region, or provide `--region` flag)
- start [openvpn client](https://openvpn.net/client/)
- import profile (click plus sign at the bottom right)
  - select "upload file" and select file `<region>-aws-vpn-<name>.ovpn` from home directory
  - connect
- delete VPN `aws-vpn delete` (same as with other commands, region can be provided via `--region` flag)

## debug

### cleanup failures

In case when cleanup/delete fails:
- check if instance was deleted
- after the instance is deleted, check and delete security group
- check and delete (if needed) instance profile (replace `<name>`)
    - `aws iam get-instance-profile --instance-profile-name aws-vpn-<name>`
    - `aws iam remove-role-from-instance-profile --instance-profile-name aws-vpn-<name> --role-name aws-vpn-<name>-<region>`
    - `aws iam delete-instance-profile --instance-profile-name aws-vpn-<name>`
- check and delete secrets (replace `<name>`, and `...` with `client.crt`, `client.key`, `ca.crt` and `ta.key`)
  - `aws secretsmanager delete-secret --secret-id /vpn/aws-vpn-<name>/secrets/... --force-delete-without-recovery`

### debug on EC2 instance

EC2 instance has SSH disabled, but user can connect via SSM.
- start openvpn service `sudo systemctl start openvpn-server@server.service` 
- openvpn service status `sudo systemctl status openvpn-server@server.service`
- logs:
  - service `sudo journalctl -xeu openvpn-server@server.service`
  - openvpn `sudo cat /etc/openvpn/server/openvpn-status.log`
  - cloud init `sudo cat /var/log/cloud-init.log`
  - cloud init output `sudo cat /var/log/cloud-init-output.log`
  - user data `curl http://169.254.169.254/latest/user-data`
- sample server config `/usr/share/doc/openvpn/sample/sample-config-files/server.conf`
- sample client config `/usr/share/doc/openvpn/sample/sample-config-files/client.conf`
