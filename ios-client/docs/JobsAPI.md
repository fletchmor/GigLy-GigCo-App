# JobsAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**jobsGet**](JobsAPI.md#jobsget) | **GET** /jobs | Get jobs list


# **jobsGet**
```swift
    open class func jobsGet(page: Int? = nil, limit: Int? = nil, status: String? = nil, location: String? = nil, completion: @escaping (_ data: [String: AnyCodable]?, _ error: Error?) -> Void)
```

Get jobs list

Get a list of jobs with optional filters and pagination

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import GigCoAPI

let page = 987 // Int | Page number (optional) (default to 1)
let limit = 987 // Int | Items per page (optional) (default to 10)
let status = "status_example" // String | Job status filter (optional)
let location = "location_example" // String | Location filter (optional)

// Get jobs list
JobsAPI.jobsGet(page: page, limit: limit, status: status, location: location) { (response, error) in
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

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **Int** | Page number | [optional] [default to 1]
 **limit** | **Int** | Items per page | [optional] [default to 10]
 **status** | **String** | Job status filter | [optional] 
 **location** | **String** | Location filter | [optional] 

### Return type

**[String: AnyCodable]**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

