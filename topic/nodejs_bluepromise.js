const bluePromise = require('bluebird');

function test() {
    promiseParams = []; // params参数数组
    bluePromise.map(
            promiseParams,
            p => toDoFunc(p),
            { concurrency: 10 }
    ).then(() => console.log('done'))

}

function toDoFunc(params) {
    //TODO
}
