npm install --save request

//get
var request = require('request');
request('您的请求url', function (error, response, body) {
  if (!error && response.statusCode == 200) {
    console.log(body) // 请求成功的处理逻辑
  }
});

//post
var request = require('request');
var url="请求url";
var requestData="上送的数据";
request({
    url: url,
    method: "POST",
    json: true,
    headers: {
        "content-type": "application/json",
    },
    body: JSON.stringify(requestData)
}, function(error, response, body) {
    if (!error && response.statusCode == 200) {
        console.log(body) // 请求成功的处理逻辑
    }
});

//form
request.post({url:'', form:{key:'value'}}, function(error, response, body) {
    if (!error && response.statusCode == 200) {
       console.log(body) // 请求成功的处理逻辑
    }
})
