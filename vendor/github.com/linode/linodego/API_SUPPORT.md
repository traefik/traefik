# API Support

## Linodes

- `/linode/instances`
  - [x] `GET`
  - [X] `POST`
- `/linode/instances/$id`
  - [x] `GET`
  - [X] `PUT`
  - [X] `DELETE`
- `/linode/instances/$id/boot`
  - [x] `POST`
- `/linode/instances/$id/clone`
  - [x] `POST`
- `/linode/instances/$id/mutate`
  - [X] `POST`
- `/linode/instances/$id/reboot`
  - [x] `POST`
- `/linode/instances/$id/rebuild`
  - [X] `POST`
- `/linode/instances/$id/rescue`
  - [X] `POST`
- `/linode/instances/$id/resize`
  - [x] `POST`
- `/linode/instances/$id/shutdown`
  - [x] `POST`
- `/linode/instances/$id/volumes`
  - [X] `GET`

### Backups

- `/linode/instances/$id/backups`
  - [X] `GET`
  - [ ] `POST`
- `/linode/instances/$id/backups/$id/restore`
  - [ ] `POST`
- `/linode/instances/$id/backups/cancel`
  - [ ] `POST`
- `/linode/instances/$id/backups/enable`
  - [ ] `POST`

### Configs

- `/linode/instances/$id/configs`
  - [X] `GET`
  - [X] `POST`
- `/linode/instances/$id/configs/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`

### Disks

- `/linode/instances/$id/disks`
  - [X] `GET`
  - [X] `POST`
- `/linode/instances/$id/disks/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `POST`
  - [X] `DELETE`
- `/linode/instances/$id/disks/$id/password`
  - [ ] `POST`
- `/linode/instances/$id/disks/$id/resize`
  - [X] `POST`

### IPs

- `/linode/instances/$id/ips`
  - [ ] `GET`
  - [ ] `POST`
- `/linode/instances/$id/ips/$ip_address`
  - [ ] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`
- `/linode/instances/$id/ips/sharing`
  - [ ] `POST`

### Kernels

- `/linode/kernels`
  - [X] `GET`
- `/linode/kernels/$id`
  - [X] `GET`

### StackScripts

- `/linode/stackscripts`
  - [x] `GET`
  - [X] `POST`
- `/linode/stackscripts/$id`
  - [x] `GET`
  - [X] `PUT`
  - [X] `DELETE`

### Stats

- `/linode/instances/$id/stats`
  - [ ] `GET`
- `/linode/instances/$id/stats/$year/$month`
  - [ ] `GET`

### Types

- `/linode/types`
  - [X] `GET`
- `/linode/types/$id`
  - [X] `GET`

## Domains

- `/domains`
  - [X] `GET`
  - [X] `POST`
- `/domains/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`
- `/domains/$id/clone`
  - [ ] `POST`
- `/domains/$id/records`
  - [X] `GET`
  - [X] `POST`
- `/domains/$id/records/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`

## Longview

- `/longview/clients`
  - [X] `GET`
  - [ ] `POST`
- `/longview/clients/$id`
  - [X] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`

### Subscriptions

- `/longview/subscriptions`
  - [ ] `GET`
- `/longview/subscriptions/$id`
  - [ ] `GET`

### NodeBalancers

- `/nodebalancers`
  - [X] `GET`
  - [X] `POST`
- `/nodebalancers/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`

### NodeBalancer Configs

- `/nodebalancers/$id/configs`
  - [X] `GET`
  - [X] `POST`
- `/nodebalancers/$id/configs/$id`
  - [X] `GET`
  - [X] `DELETE`
- `/nodebalancers/$id/configs/$id/nodes`
  - [X] `GET`
  - [X] `POST`
- `/nodebalancers/$id/configs/$id/nodes/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`
- `/nodebalancers/$id/configs/$id/rebuild`
  - [X] `POST`

## Networking

- `/networking/ip-assign`
  - [ ] `POST`
- `/networking/ips`
  - [X] `GET`
  - [ ] `POST`
- `/networking/ips/$address`
  - [X] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`

### IPv6

- `/networking/ips`
  - [X] `GET`
- `/networking/ips/$address`
  - [X] `GET`
  - [ ] `PUT`
- /networking/ipv6/ranges
  - [X] `GET`
- /networking/ipv6/pools
  - [X] `GET`

## Regions

- `/regions`
  - [x] `GET`
- `/regions/$id`
  - [x] `GET`

## Support

- `/support/tickets`
  - [X] `GET`
  - [ ] `POST`
- `/support/tickets/$id`
  - [X] `GET`
- `/support/tickets/$id/attachments`
  - [ ] `POST`
- `/support/tickets/$id/replies`
  - [ ] `GET`
  - [ ] `POST`

## Account

### Events

- `/account/events`
  - [X] `GET`
- `/account/events/$id`
  - [X] `GET`
- `/account/events/$id/read`
  - [X] `POST`
- `/account/events/$id/seen`
  - [X] `POST`

### Invoices

- `/account/invoices/`
  - [X] `GET`
- `/account/invoices/$id`
  - [X] `GET`
- `/account/invoices/$id/items`
  - [X] `GET`

### Notifications

- `/account/notifications`
  - [X] `GET`

### OAuth Clients

- `/account/oauth-clients`
  - [ ] `GET`
  - [ ] `POST`
- `/account/oauth-clients/$id`
  - [ ] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`
- `/account/oauth-clients/$id/reset_secret`
  - [ ] `POST`
- `/account/oauth-clients/$id/thumbnail`
  - [ ] `GET`
  - [ ] `PUT`

### Payments

- `/account/payments`
  - [ ] `GET`
  - [ ] `POST`
- `/account/payments/$id`
  - [ ] `GET`
- `/account/payments/paypal`
  - [ ] `GET`
- `/account/payments/paypal/execute`
  - [ ] `POST`

### Settings

- `/account/settings`
  - [ ] `GET`
  - [ ] `PUT`

### Users

- `/account/users`
  - [ ] `GET`
  - [ ] `POST`
- `/account/users/$username`
  - [ ] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`
- `/account/users/$username/grants`
  - [ ] `GET`
  - [ ] `PUT`
- `/account/users/$username/password`
  - [ ] `POST`

## Profile

### Personalized User Settings

- `/profile`
  - [ ] `GET`
  - [ ] `PUT`

### Granted OAuth Apps

- `/profile/apps`
  - [ ] `GET`
- `/profile/apps/$id`
  - [ ] `GET`
  - [ ] `DELETE`

### Grants to Linode Resources

- `/profile/grants`
  - [ ] `GET`

### SSH Keys

- `/profile/sshkeys`
  - [x] `GET`
  - [x] `POST`
- `/profile/sshkeys/$id`
  - [x] `GET`
  - [x] `PUT`
  - [x] `DELETE`
  
### Two-Factor

- `/profile/tfa-disable`
  - [ ] `POST`
- `/profile/tfa-enable`
  - [ ] `POST`
- `/profile/tfa-enable-confirm`
  - [ ] `POST`

### Personal Access API Tokens

- `/profile/tokens`
  - [ ] `GET`
  - [ ] `POST`
- `/profile/tokens/$id`
  - [ ] `GET`
  - [ ] `PUT`
  - [ ] `DELETE`

## Images

- `/images`
  - [x] `GET`
- `/images/$id`
  - [x] `GET`
  - [X] `POST`
  - [X] `PUT`
  - [X] `DELETE`

## Volumes

- `/volumes`
  - [X] `GET`
  - [X] `POST`
- `/volumes/$id`
  - [X] `GET`
  - [X] `PUT`
  - [X] `DELETE`
- `/volumes/$id/attach`
  - [X] `POST`
- `/volumes/$id/clone`
  - [X] `POST`
- `/volumes/$id/detach`
  - [X] `POST`
- `/volumes/$id/resize`
  - [X] `POST`
