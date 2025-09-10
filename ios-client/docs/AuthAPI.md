# AuthAPI

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**authLoginPost**](AuthAPI.md#authloginpost) | **POST** /auth/login | Login user
[**authRegisterPost**](AuthAPI.md#authregisterpost) | **POST** /auth/register | Register a new user


# **authLoginPost**
```swift
    open class func authLoginPost(credentials: ApiLoginRequest, completion: @escaping (_ data: ApiLoginResponse?, _ error: Error?) -> Void)
```

Login user

Authenticate user and return JWT token

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import GigCoAPI

let credentials = api.LoginRequest(email: "email_example", password: "password_example") // ApiLoginRequest | Login credentials

// Login user
AuthAPI.authLoginPost(credentials: credentials) { (response, error) in
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
 **credentials** | [**ApiLoginRequest**](ApiLoginRequest.md) | Login credentials | 

### Return type

[**ApiLoginResponse**](ApiLoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **authRegisterPost**
```swift
    open class func authRegisterPost(user: ApiRegisterRequest, completion: @escaping (_ data: ApiRegisterResponse?, _ error: Error?) -> Void)
```

Register a new user

Register a new user in the system with email, name and role

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import GigCoAPI

let user = api.RegisterRequest(address: "address_example", availability: "availability_example", email: "email_example", latitude: 123, longitude: 123, name: "name_example", phone: "phone_example", placeId: "placeId_example", role: "role_example", skills: ["skills_example"]) // ApiRegisterRequest | Registration data

// Register a new user
AuthAPI.authRegisterPost(user: user) { (response, error) in
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
 **user** | [**ApiRegisterRequest**](ApiRegisterRequest.md) | Registration data | 

### Return type

[**ApiRegisterResponse**](ApiRegisterResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

