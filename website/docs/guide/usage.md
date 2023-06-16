---
title: Using vince 
----

Vince uses environment variables and commandline arguments for configuration  
[Configuration](../guide/config) for full list of options and their meaning.

In this guide we will setup a simple vince instance.


## step 0 - configure

```
vince config >.env
```
::: details .env file contents
<<< @/guide/files/env.sh
:::

This will generate a default configuration that we can customize for our use case

We need to decide where our data will live. For our case lets use `.vince`

```
mkdir .vince
```

**Setup administrator**

Edit `.env`

```sh
# allows creating a user and api key on startup.
export  VINCE_ENABLE_BOOTSTRAP="false" // [!code --]
export  VINCE_ENABLE_BOOTSTRAP="true" // [!code ++]
# Full name of the user to bootstrap.
export  VINCE_BOOTSTRAP_NAME="Jane Doe" // [!code ++]
export  VINCE_BOOTSTRAP_NAME="" // [!code --]
# Email address of the user to bootstrap.
export  VINCE_BOOTSTRAP_EMAIL="janedoe@example.com" // [!code ++]
export  VINCE_BOOTSTRAP_EMAIL="" // [!code --]
# Password of the user to bootstrap.
export  VINCE_BOOTSTRAP_PASSWORD="my_secret_password" // [!code ++]
export  VINCE_BOOTSTRAP_PASSWORD="" // [!code --]
```

**Load .env in our shell**
```sh
source .env
```

**Start vince**
```sh
vince
```

This will start vince on port `8080` you can visit http://localhost:8080 to login
with the email and password you set above.

## Step 1 -

There is no step 1. Congratulation , you have successful deployed vince. What
follows after this is just modifying `.env` to your heart's desire , load the
`.env` file  on the shell and run `vince`.

Thanks , and Welcome.
