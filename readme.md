# Engineering task - simple rate limiter

```
Implement web service that acts as third party rate limiter service.
```
## Specifications

The service should:
1. Accept following arguments during the startup (command line args):
threshold - Max number of requests per URL within a pre-defined time period (ttl).
ttl - The pre-defined time period in which URL visits will be counted.
2. Expose endpoint to report URL visit in the following format:

```
POST /report
Content-Type: application/json
{
"url": "http://www.sample.com"
}
```

3. The response to each request should be "block" true/false - depends if the number of times the URL was
reported passed the threshold:

```
{
"block": true // if it passed the threshold
}
```

4. Track the number of times a URL was reported within the ttl period.
5. Each URL should be hashed in order to reduce memory usage.
6. The rate-limit per URL should be reset after ttl defined in ms.

## Notes


Do not use external services - it should be implemented in a single process by your code.
Implement the above in GO.
Make sure service has no resource leaks.
Logging will be appreciated.
Clean code will be appreciated.

