# HealthAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**healthGet**](HealthAPI.md#healthget) | **GET** /health | Check API health status


# **healthGet**
```swift
    open class func healthGet(completion: @escaping (_ data: [String: AnyCodable]?, _ error: Error?) -> Void)
```

Check API health status

Check if the API and database are healthy

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import GigCoAPI


// Check API health status
HealthAPI.healthGet() { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

**[String: AnyCodable]**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

