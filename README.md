raspberry
=========

An open source gateway by go.

The API Gateway sits between your API and the wider internet, essentially it acts as a reverse proxy which filters your traffic as it comes towards your API.

Features
---------

- Request throttling
- Quota periods
- Path white-listing
- Simple, secure REST API for adding, deleting and quering API keys
- Analytics gathering (to CSV or MongoDB)

Configuration
-------------

Configuration is handled with a JSON file, this will be generated for you on first run, alternatively you can specify it with the `--conf=<>filename` command line flag.

The configuration paramaters are pretty straightforward and are explained below:

    {
        "listen_path":"/gateway",
        "listen_port":8080,
        "target_url":"http://localhost:8080/api",
        "secret":"352d20ee67be67f6340b4c0605b044b7",
        "template_path":"templates",
        "auth_header_name":"lz-authentication-token",
        "storage":{
            "type":"redis",
            "host":"localhsot",
            "port":6379,
            "username":"",
            "password":""
        },
        "exclude_paths": [
            "/v1/sessions"
        ],
        "enable_analytics", true,
        "analytics_config": {
            "type": "mongo",
            "csv_dir": "/vagrant/logs/",
            "purge_delay": 3600,
            "mongo_url": "mongodb://analytics:forgetmenot!@candidate.19.mongolayer.com:10167/loadzen_elastic",
            "mongo_db_name": "Loadzen_elastics",
            "mongo_collection": "raspberry_analytics"
        }
    }

### listen_path
The path to intercept requests on, in this case we specified `/gateway`, this is the URL that Raspberry will listen on and apply filtering and quota rules to. For example, if your API documentation state that a  resource is at `/gateway/widgets` then Raspberry will intercept and apply filtering. All other URL's will return a `404`. It is recommended to set this to `/` and use a web server to manage path-level reverse proxying (e.g. Nginx)

### listen_port
Pretty obvious - the port to bind to

### target_url
The URl to reverse proxy - if we have set the `listen_path` to `/api`, then traffic going to `http://api.domain.com/api/widgets` will be proxied to `http://respberry.io/api/widgets`.

### secret
This value is required as part of the Raspberry API call, if you want use any key management api's this secret will need to be sent along as part of the request headers as `x-raspberry-authorisation`.

### template_path
The path where to find templates, defaults to `templates` in the current directory. Only one template exists: `error.json`, this does not nedd to be json file, it can be xml - just the filename should not be changed! It follows the Go template syntax and can be used to serve error messages in a standard format as your API requires.

### auth_header_name
The authentication header that Raspberry will use to find the API key to access your API. Currently only header values are supported.

### storage
Details for the data store of the API Keys which Raspberry uses, two options are possible: `redis` and `memory`, if `memory` is used then keys will be stored in RAM - this is not recommended but handy for testing. `redis` is the recommended setting and requires a Redis installation and the remaining section details to be filled in.

### exclude_paths
If you have API paths that do not require authorisation, describe thme here, these will be proxied without auth or quota checks.

### analytics_config
The configuration of how to store analytics. Two modes are supported for the `type` key: `csv` and `mongo`, selecting the `csv` type will cause Raspberry to purge the access data from Redis to disk (in the `csv_dir` directory - this must be an absolute path) at the rate specified by `purge_delay` (in seconds). The `mongo` type will store this data in a MongoDB instance of your chosing, ensure the details are correct in order to connect.

Recommended Deployment
-----------------------

It is our opinion that the safest way to set up Raspberry is is behind another web server such as NginX, as it should act as a part of your ingrastructure stack, and not be front-line. A typical setup would involve:

1. NGinx acting as the web server
2. API requests (via host or path) are reverse-proxied upstream to multiple Raspberry processes, which is listening on a non-public port
3. Raspberry is configured to target your API application and listen on `/`

Raspberry can be load-balanced the same way any other web server can, so you can have multiple instances running on different ports.

To run multiple procs we suggest using a tool such as supervisored.

Starting raspberry
------------------

Starting Raspberry is very simple:

    > ./raspberry [--conf=raspberry.conf] [--port=PORT]

The `--conf` flag is optional, Raspberry will create a configuration file if it can't find one. Stating `--port` will override the port set in the configuration file.