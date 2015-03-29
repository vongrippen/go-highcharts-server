# go-highcharts-server

Tiny go webapp for generating highcharts charts on the server for including in emails, PDFs, etc.
Uses highcharts-convert.js. [Link](http://www.highcharts.com/component/content/article/2-news/52-serverside-generated-charts)

The following are accepted as HTTP parameters:

* `input`  A javascript object with the chart configuration (same as `-infile`)
* `scale`  The zoomFactor of the page used to generate the chart
* `width`  The exact pixel width of the chart
* `constr`  The constructor name. Can be either _Chart_ or _StockChart_ (if Highstock is included)
* `callback`  A javascript callback function to be executed on chart load
* `type`  The file type to be returned. Can be one of _png_ (default), _jpg_, _pdf_, or _svg_

There are a few environment variables that can be set to control how the webapp runs:

* `HTTP_BASIC_USERNAME`  An HTTP Basic Auth username can be required (also requires a password)
* `HTTP_BASIC_PASSWORD`  An HTTP Basic Auth password can be required (also requires a username)
* `IP`  The IP address to listen on (defaults to 0.0.0.0)
* `PORT`  The port to listen on for HTTP requests (defaults to 8080)
* `KEEPALIVE_URL`  A URL to ping regularly to keep the app marked as active (useful for Heroku) This should be set to http://_your site_/ping

Requires go 1.0 (or higher) and phantomjs.

## Getting Started Locally

Clone the project

```bash
git clone git://github.com/vongrippen/go-highcharts-server.git
```

Install phantomjs
```bash
gem install brew
brew install phantomjs
```

Start up the server
```bash
go run
```

Generate a chart using curl
```bash
HIGHCHART_OBJECT=`cat ./spec/fixtures/input.json`
curl -X POST -d "input=$HIGHCHART_OBJECT" http://localhost:8080/ -o ./chart.png
```

Or optionally with width:
```bash
curl -X POST -d "input=$HIGHCHART_OBJECT&width=900" http://localhost:8080/ -o ./chart.png
```


## Deploying to Heroku
  
This project is made to deploy to heroku.  It is using the [heroku-buildingpack-multi](https://github.com/ddollar/heroku-buildpack-multi)
depending including both phantomjs and go.  Just create a new project within heroku 
and push it like you know how.

## Calling API from ruby

Using [httparty](https://github.com/jnunemaker/httparty):
```ruby
require 'httparty'

chart_object_js = '{series: [{
    type: 'pie',
    name: 'Slices',
    data: [1,2,3,4,5]
  }]}'
response = HTTParty.post('http://localhost:8080/', body: {input: chart_object_js, width:550})
File.open('./chart.png', 'wb'){ |file| file << response.body }
```



