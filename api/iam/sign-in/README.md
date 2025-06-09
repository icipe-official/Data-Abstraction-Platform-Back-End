# Sign in using OpenID Access and Refresh Token

Return the iam credential as well as the platform's encrypted access and refresh token based on the openid information.

Requires OpenID Access and Refresh Token as well as their expiry time.

Go [here](../../../../docs/setup/keycloak/api/get-token/README.md) to generate the `json` file that contains this information and copy its content into `input_token.json`.

## Request

```http
@hostname=
@openidaccesstoken=
@openidaccesstokenexpiresin=
@openidrefreshtoken=
@openidrefreshtokenexpiresin=

GET /iam/sign-ing HTTP/1.1
Host: {{hostname}}
Accept: application/json
OpenID-Access-Token: {{openidaccesstoken}}
OpenID-Access-Token-Expires-In: {{openidaccesstokenexpiresin}}
OpenID-Refresh-Token: {{openidrefreshtoken}}
OpenID-Refresh-Token-Expires-In: {{openidrefreshtokenexpiresin}}
```

## Response

Sample Json response [here](./sample_response.json).

## Request execution

To run the request, the following pre-requisite may be met:

1. Setup Enviornment variables - Refer to [Setting up Environment Variables](../README.md#environment-variables) for setting up env variables. In this case execute the instructions for the env in the [current folder](./env.sh.template).

### Flags

<table>
    <thead>
        <th>Flag</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>--verbose</td>
            <td>Log request information.</td>
        </tr>
        <tr>
            <td>--output-json</td>
            <td>Output json result of openid configuration in <code>output.json</code> in the script directory.</td>
        </tr>
    </tbody>
</table>
