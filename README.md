# webscrot
Simple tool that concurrently launches given amount of browser instances on virtual displays, sleeps for specified time and takes a screenshot of the browser (page) until all URLs are exhausted.

Input file is a simple JSON array of strings (URLs) `["http://github.com/adnanh/webscrot", "http://news.ycombinator.com"]`

# How to get it
Make sure you've set up your `$GOPATH` environment variable properly and then run 
```
go get github.com/adnanh/webscrot
```

# Command line flags
```
  -workers int
      number of concurrent workers (default 1)
  -fileprefix string
    	string to prefix the output filename
  -filesuffix string
    	string to suffix the output filename
  -width int
    	screen width (default 1024)
  -height int
    	screen height (default 768)
  -outputpath string
    	path where the screenshots should be saved (default "./")
  -timeout int
    	number of milliseconds to wait before taking a screenshot (default 5000)
  -urlfile string
    	path to the JSON file containg an array of urls (default "url.json")
  -urlprefix string
    	string to prefix the urls from JSON file
```

# Requirements
- Xvfb
- ratpoison
- ImageMagick
- midori

## Get them via package manager
### ArchLinux
`sudo pacman -Sy xorg-server-xvfb ratpoison imagemagick midori`

### Fedora 22
`sudo dnf install xorg-x11-server-Xvfb ratpoison midori ImageMagick`

# License
The MIT License (MIT)

Copyright (c) 2015 Adnan Hajdarevic <adnanh@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
