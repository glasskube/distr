# Distr SDK

You can install the Distr SDK for JavaScript from [npmjs.org](https://npmjs.org/):

```shell
npm install --save @glasskube/distr-sdk
```

Conceptually, the SDK is divided into two parts:

- A high-level service called `DistrService`, which provides a simplified interface for interacting with the Distr API.
- A low-level client called `Client`, which provides a more direct interface for interacting with the Distr API.

In order to connect to the Distr API, you have to create a Personal Access Token (PAT) in the Distr web interface.
Optionally, you can specify the URL of the Distr API you want to communicate with. It defaults to `https://app.distr.sh/api/v1`.

```typescript
import {DistrService} from '@glasskube/distr-sdk';
const service = new DistrService({
  // to use your selfhosted instance, set apiBase: 'https://selfhosted-instance.company/api/v1',
  apiKey: '<your-personal-access-token-here>',
});
// do something with the service
```

The [src/examples](https://github.com/distr-sh/distr/tree/main/sdk/js/src/examples) directory contains examples of how to use the SDK.

See the [docs](https://github.com/distr-sh/distr/tree/main/sdk/js/docs/README.md) for more information.
