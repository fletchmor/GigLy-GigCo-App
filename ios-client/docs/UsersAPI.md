# UsersAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**usersProfileGet**](UsersAPI.md#usersprofileget) | **GET** /users/profile | Get current user profile


# **usersProfileGet**
```swift
    open class func usersProfileGet(completion: @escaping (_ data: ModelUser?, _ error: Error?) -> Void)
```

Get current user profile

Get the profile of the currently authenticated user

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import GigCoAPI


// Get current user profile
UsersAPI.usersProfileGet() { (response, error) in
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

[**ModelUser**](ModelUser.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

