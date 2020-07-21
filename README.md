# Get Qmail Forwarding Rules From Webdav Storage

Our mail system lets users create as many qmail aliases as they want, most of them forwarding to a single POP or IMAP store somewhere else.

Delivery instructions for a user's addresses live on a webdav share accessed by stateless MX/forwarding hosts, and by the user, who can mount the shares locally.

When users mount their forwarding directory, they can edit files to control exactly what happens to mail to any of their addresses. The current version of this utility allows users to drop or bounce messages, forward to any address(es), or bounce messages unless there is a certain string in the subject.

Email addresses can optionally be created on first use, letting users make up addresses on the fly without having to set them up ahead of time.
If an address starts receiving spam, it can be blocked without affecting other addresses.

## How it works

The delivery agent `qdeliver` is placed in a `.qmail` default file and passed the recipient address on the command line.
Something like this in `~alias/.qmail-example-default`:

```
|/path/to/qdeliver "$EXT2" "$HOST"
```

`qdeliver` extracts the first "-" delimited token from the first argument and uses that plus the second argument as keys into a `users.json` configuration file that looks like this:

```
{
    "version": 1,
    "accounts": [
        {
            "owner": "me",
            "domain": "example.com",
            "url": "https://webdav.example.com/example.com/",
            "login": "me",
            "password": "WebDav-Pa55w0rd",
            "notify": true
        }
    ]
}
```

So mail to `me@example.com` and `me` plus extensions will be controlled by files on the webdav server under the `example.com` directory.

Files in that directory are text files named after the localpart of the address, with a `.txt` extension to make it easier for editing applications to see them.

Here is an example `me.txt` file:

```
forward me@pop.example.com
```

Here is an example `me-foo.txt` file where the address (`me-foo@example.com`) has been disabled because of spam:

```
drop
```

The delivery instructions are not normal `.qmail` instructions.
They are deliberately very limited.

If the address file doesn't exist, but a `default.txt` file does, then `default.txt` will be copied to the address file and then executed as if it already existed.
So address files can be automatically generated.

## How to build and install

First make sure go is installed, then clone this repo and do this:

```
cd qdeliver
go test ./...
go build ./cmd/qdeliver
```

The easiest way to install is to copy `qdeliver` and the scripts under `scripts/` to `~alias`.
Then you don't have to use extra command line options to specify the location of the scripts.
But you can put these files anywhere you want.
Look at the `qdeliver` man page for more details.

## How to setup the webdav contents

A webdav directory holds files for a single base address and all of its extension addresses.
Within this directory are files named after the localpart of the address with a `.txt` filename extension.

For example, the directory might hold the following files:

```
default.txt
joe.txt
joe-amazon.txt
joe-ebay.txt
```

Include the path to this directory in the `url` of `users.json`.
If you create a `default.txt`, the files for new addresses will automatically be created.
If not, mail to addresses without an address file will bounce.
If you don't include the base address file (`joe.txt` above), then mail to the base address will bounce.

Address files are text files with one instruction per line.
Comments and empty lines are skipped.
Strings with spaces can be enclosed in single or double quotes.
Double quotes allow \\-escaped characters to be embedded in the string.

The following instructions are currently recognized:

|keyword|arguments|meaning|
|---|---|---|
| forward | address(es) to forward to | forward message to one or more space-separated addresses
| bounce | optional string | bounce message; if string is given, it will be included in the bounce message
| drop | | eat the message; don't forward, don't bounce
| match-subject | string | bounce message unless string is found in subject

Anything causes the delivery to be deferred.

Examples:

```
forward some.body@example.com another@pop.example.com
```
```
bounce 'Changed address to foo@example.com'
```
```
match-subject 'code-word'
forward me@example.com
````

## How to configure qdeliver

The main configuration is `users.json`.
See above for the structure of this file.
Create an `accounts` element for every basename+domain combination you want to handle.

Put `users.json` in the qdeliver execution directory (eg `/var/qmail/alias`), or use the `--db` command line flag to specify a different location.

## How to configure qmail

There are many ways to configure qmail and `qdeliver`.
This is one example.

### rcpthosts

Add `qdeliver` domains to `rcpthosts`.

### virtualdomains

Add `qdeliver` domains to `virtualdomains`, something like this:
```
myvirtualdomain.com:alias-virt
```
This will cause `~alias/.qmail-virt-default` to control mail to that domain.

### ~alias.qmail-*

Create the appropriate `.qmail-something-default` file in `~alias` with contents like this:
```
|./qdeliver "$EXT2" "$HOST"
```
You might have to use a different `$EXT?` if your setup is different.

### Restart qmail-send

Something like:
```
svc -t /service/qmail-send
```

### Test delivery locally

Use local mail to send mail to an address controlled by `qdeliver` and where you have the webdav contents set up.

### MX record

Set up an MX record for your controlled domain, pointing to your server.