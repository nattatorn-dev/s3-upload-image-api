# img

pure go image rendering server. accepts a POST, stores the file on s3, and performs various manipulations on the images via GET string params

**example upload**

```
curl -i -F file=@photo2.jpg http://localhost:8080/upload
```

## API

- `action`: the action to perform. currently supports:
  crop: currently only cropping form the middle is supported
- `w`: width in pixels of the cropped image
- `h`: height in pixels of the cropped image

**example request**

```
http://localhost:8080/img/671e3346e405b99441bf4f0de7abc4dd?action=thumbnail&w=500&h=500
```

## Config

Use the `config.toml` file to config various variables of the server.

## AWS creds

The AWS package expects a credentail file located at `~/.aws/credentials` of ENV variables setup. [Additional details here](https://github.com/aws/aws-sdk-go).

```
$filepath = 'image.jpg',
$endpoint = 'http://192.168.6.199:8080/process',
$path = '/krneka-podmapa/',
$image_id = 1000,
$slug = 'image-'.uniqid(),
$schema = 'via'
```
