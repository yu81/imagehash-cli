# imagehash-cli
image hash calculation CLI with https://github.com/corona10/goimagehash

## Install
```
go install github.com/yu81/imagehash-cli
```

## Usage
### single image hash calculation
```
# Perception Hash
imagehash-cli -p path_to_image

# Average Hash
imagehash-cli -a path_to_image

# Difference Hash
imagehash-cli -d path_to_image

# Wavelet Hash (requires very large memory)
imagehash-cli -w path_to_image
```
output
```
14648015642502059055
```

# Distance between two images

```
# You can choose hash algorithm as the same as single image hash calculation's ones.
imagehash_cli -p path_to_image1 path_to_image2 
```
output
```
13754156118075818639 14648015642502059055 28
```