# UT2 Utility (ut2u) 

`ut2u` is a utility to assist in UT2004 development and server maintenance
tasks.

With `ut2u`, you can:

* Generate a manifest containing information for the UT2004 packages on your
  server.
* Automatically compress your UT2004 packages and synchronize it to S3 object
  storage.


## Redirect

The `ut2u redirect` command can upload your packages to S3 object storage,
automatically compressing them for you.


### The `RedirectToURL` Option

The `RedirectToURL` option under `IpDrv.HTTPDownload` has template support for
the following:

* `%guid%` - Package GUID
* `%ext%` - Package extension
* `%lcext%` - Package extension lowercase
* `%ucext%` - Package extension uppercase
* `%file%` - Package filename
* `%lcfile%` - Package filename lowerase
* `%ucfile%` - Package filename uppercase

Using solely the filename may cause conflicts when game mods and mutators are
re-released under the same name. To help, `ut2u redirect` follows the following
pattern:

```
[IpDrv.HTTPDownload]
RedirectToURL=http://redirect.example.com/%file%/%guid%
```

NOTE: The `%file%` substitution will automatically append `.uz2` to the package
name if compression is enabled on the server. Learned this the hard way.


### Upload

`ut2u redirect upload` uploads Unreal packages to a redirect server, compressing
it in the process.

```
export AWS_ACCESS_KEY_ID=xxxxx
export AWS_SECRET_ACCESS_KEY=xxxxx
export AWS_DEFAULT_REGION=us-east-1
```

It is expected you have a bucket with permissions for `ListBucket`, `GetObject`,
and `PutObject`.

```
ut2u redirect upload -b my.bucket -p ut2-redirect/ Maps/DM-Rankin.ut2
```

The DM-Rankin.ut2 package is now uploaded to
`my.bucket/ut2-redirect/DM-Rankin.ut2/<GUID>.uz2`


### Sync

`ut2u redirect sync` reads your server's `UT2004.ini` and searches for Unreal
packages. Packages are checked to see if they exist on the redirect server and
uploaded if not.

Configure your AWS credentials similarily to the `upload` command.

```
ut2u redirect sync -b my.bucket -p ut2-redirect/ System/UT2004.ini
```
