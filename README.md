# UT2 Utility (ut2u) 

`ut2u` is a utility to assist in UT2004 development and server maintenance
tasks.

With `ut2u`, you can:

* Query servers.
* Extract package information including: GUID, dependencies, and
  checksums.
* Compress/Decompress packages.
* Search for dependents of a package.
* Generate a manifest containing information for the UT2004 packages on your
  server.
* Automatically compress your UT2004 packages and synchronize it to S3 object
  storage.


## Query

`ut2u query` queries a given server. The server must be of the form
`ip:port`, with port being the game port.

```console
$ ut2u query chi-1.staging.kokuei.dev:7777
UTComp Duels [ut2.chi-1.staging.kokuei.dev/1]
  Game: xDeathMatch
  Map: DM-DE-Ironic-FE
  Players: 0/2
  Rules:
    ServerMode: dedicated
    AdminName: kokuei
    AdminEmail: hello@kokuei.dev
    ServerVersion: 3369
    GameStats: False
    MaxSpectators: 24
    MapVoting: true
    KickVoting: false
    Mutator: MutAntiTCCFinal
    Tick Rate: 58.97 / 60.00 max.
    Mutator: iTSFake
    Mutator: MutReplace
    Mutator: MutUseLightning
    Mutator: MutNoAdrenaline
    Mutator: MutNoDoubleDamage
    Mutator: MutNoSuperWeapon
    Mutator: MutUTComp
    UTComp Version: 1.8b-K4
    Enhanced Netcode: True
    Mutator: Tickrate
    MinPlayers: 2
    EndTimeDelay: 4.00
    GoalScore: 0
    TimeLimit: 15
    Translocator: False
    WeaponStay: False
    ForceRespawn: True
    mutator: DMMutator
```

You may pass in multiple servers and they will be queried simultaneously.


## Packages

`ut2u package` offers various commands related to UT2 packages.


### Compression/Decompression

`ut2u` supports package compression and decompress without the need for UCC.

```console
$ ut2u package compress DM-Test.ut2
DM-Test.ut2 -> DM-Test.ut2.uz2
```

```console
$ ut2u package decompress DM-Test.ut2.uz2
DM-Test.ut2.uz2 -> DM-Test.ut2
```


### Info

`ut2u package info` will return information about a package.

```console
$ ut2u package info DM-Test.ut2
Name:     DM-Test.ut2
GUID:     8BD57B014CEE4E6523AEF5BE1C6DCE89
Provides: DM-Test
Requires:
  - 2K4Chargers
  - Engine
  - UT2004Weapons
  - XGame
Checksums:
  MD5:    77329b17d456a165b5124a65ab88c209
  SHA1:   097f76bdfafd43eda12803de67ff2a9197886921
  SHA256: fd96be829e728c617808d953f57c41c67b5a5dbfdd7151a6d326b1e6da628c7b
```


### Check Dependencies

`ut2u package check-deps` verifies every package in your UT2004
installation has its dependencies met. It requires you pass in the
path to your UT2004.ini file.

```console
$ ut2u package check-deps /path/to/System/UT2004.ini
Package DanielsMeshes.usx has missing dependencies: Belt_fx
Package VehicleMeshes.usx has missing dependencies: VehicleSkins
Some packages have missing dependencies
```


### Requires

`ut2u package requires` searches your UT2004 installation for packages that
list the given package as a dependency. The package should not have an
extension.

```console
$ ut2u package requires /path/to/System/UT2004.ini DEBonusTextures
BR-DE-ElecFields.ut2
CTF-DE-ElecFields.ut2
CTF-Grendelkeep.ut2
DEBonusMeshes.usx
DM-DE-Grendelkeep.ut2
DM-DE-Ironic.ut2
DM-DE-Osiris2.ut2
ONS-Aridoom.ut2
```


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
