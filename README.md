# img

pure go image rendering server. accepts a POST, stores the file on s3, and performs various manipulations on the images via GET string params.

## API

**example upload**

```
http://localhost:8080/upload

$file = 'image.jpg',
$path = '/example/',
$image_id = 1000,
$slug = 'image',
$schema = 'via'
```

**example request**

```
http://localhost:8080/img/:id
```

## Config

Use the `config.toml` file to config various variables of the server.
