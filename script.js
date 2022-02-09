import http from 'k6/http'

import { check } from 'k6'

export let options = {
    vus: 100,
    iterations: 100000
}


export default function () {
    const method = __ENV.METHOD || ''
    let res
    switch (method.toLowerCase()){
        case 'get':
            res = http.get(`http://localhost:1234/${__ENV.ENDPOINT}`)
            break
        case 'post':
            res = http.post(
                `http://localhost:1234/${__ENV.ENDPOINT}`,
                JSON.stringify({
                    tool: 'k6.io',
                    description: 'The best developer experience for load testing'
                }),
                { headers: { 'Content-Type': 'application/json' } }
            )
            break
        default:
            throw new Error('method not support')
    }

    check(res, { 'status was 200': (r) => r.status == 200 });
}