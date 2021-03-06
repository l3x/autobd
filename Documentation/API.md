# Autobd's HTTP API

# GET /version
### Description:
Returns a JSON encoded structure describing the version of the autobd server. 
Autobd nodes use this endpoint to ensure version equality with the server.

### Arguments:
None

### Example:
```
https://host:8080/version
```
### Returns:
```
{
    "server": "0.0.4",
    "commit": "8d913b8990"
}
```

### Status:
- 200 OK: Call succeeded, returns expected json struct

# GET /index
### Description:
Returns a JSON encoded structure describing the files and directory tree on the server

### Arguments:

```
dir=<requested directory to index>
```
The directory to index

```
uuid=<registered node UUID>
```
The node requesting the index, must already be identified on the server

### Example: 

```
http://host:8080/v0/index?dir=/&uuid=a468d5d0-56b8-4b0d-be2f-08b7d612b055
```

### Returns:
```
{
    "directory1": {
      "name": "directory1",
      "size": 4096,
      "lastModified": "2016-09-26T16:46:05.468167071-06:00",
      "fileMode": 2147484141,
      "isDir": true
    },
    "directory2": {
      "name": "directory2",
      "size": 4096,
      "lastModified": "2016-09-26T16:46:07.918183525-06:00",
      "fileMode": 2147484141,
      "isDir": true
    },
    "directory3": {
      "name": "directory3",
      "size": 4096,
      "lastModified": "2016-09-28T19:13:49.163346347-06:00",
      "fileMode": 2147484141,
      "isDir": true,
      "files": {
        "directory3/file": {
          "name": "directory3/file",
          "checksum": "8d0a77a2685b1c3781de27043f71e487a0fd8472ce08917959fc6819bd32e81a636e5f817a948fa24f6f1427978dbaeb01a26a9f214aafd10ca379086bfc3ab1",
          "size": 1002,
          "lastModified": "2016-09-29T20:54:47.516394557-06:00",
          "fileMode": 420,
          "isDir": false
        }
      }
    }
  }
```
### Status:
- 200 OK: Call succeeded, returns expected json struct
- 400 Bad Request: Directory not found or directory not in request
- 500 Internal Server Error: Error while processing sync request
- 501 Unauthorized: UUID not found in node list or UUID not in request

# GET /sync
### Description:
Returns the requested file (gzip'd, if the node-side can handle it) or a directory, (tarballed and gzip'd if the node-side can handle it)


### Arguments: 

```
grab=<file or directory path> 
```
The file or directory to transfer

```
uuid=<registered node UUID>
```
The node requesting the sync, must already be identified on the server

### Example:
```
http://host:8080/v0/sync?grab=/directory3&uuid=a468d5d0-56b8-4b0d-be2f-08b7d612b055
```

### Returns:
Contents of requested directory in gzip'd format

### Status:
- 200 OK: Call succeeded, returns requested directory contents
- 400 Bad Request: Directory not found or directory not in request
- 500 Internal Server Error: Error while processing server index or index request
- 501 Unauthorized: UUID not found in node list or UUID not in request

# GET /nodes

### Description:
Returns a list of nodes currently registered with the server and their metadata, encoded in json

### Example:
```
http://host:8080/v0/nodes?uuid=a468d5d0-56b8-4b0d-be2f-08b7d612b055
```

### Returns:
```
{
  "709225b3-e8c9-44f7-9f92-cd9bace5d533": {
   "address": "127.0.0.1:43226",
   "last_online": "Saturday, 11-Feb-17 15:02:58 MST",
   "is_online": true,
   "synced": false,
   "metadata": {
    "version": "0.0.0",
    "UUID": "709225b3-e8c9-44f7-9f92-cd9bace5d533"
   }
  },
  "7a139721-3323-4b58-b6a0-2fc7c574338f": {
   "address": "127.0.0.1:43222",
   "last_online": "Saturday, 11-Feb-17 15:02:30 MST",
   "is_online": true,
   "synced": false,
   "metadata": {
    "version": "0.0.0",
    "UUID": "7a139721-3323-4b58-b6a0-2fc7c574338f"
   }
  },
  "c24506d3-0d70-4642-8208-207895b1738e": {
   "address": "127.0.0.1:43232",
   "last_online": "Saturday, 11-Feb-17 15:03:01 MST",
   "is_online": true,
   "synced": false,
   "metadata": {
    "version": "0.0.0",
    "UUID": "c24506d3-0d70-4642-8208-207895b1738e"
   }
  }
 }
```

### Status:
- 200 OK: Request succeeded, returns list of nodes currently registered with this server
- 501 Unauthorized: UUID not found in node list or UUID not in request


# POST /heartbeat
### Description:
Updates the node's status on the server

### Arguments:
A NodeHeartbeat struct, populated with the node's UUID and synced status, encoded in json

### Example:
```
http://host:8080/v0/heartbeat?uuid=a468d5d0-56b8-4b0d-be2f-08b7d612b055&synced=true
```

### Returns:
Nothing

### Status:
- 200 OK: Node with UUID status is updated
- 500 Internal Server Error: Error while processing heartbeat request or error while updating node status
- 501 Unauthorized: UUID in request not recognized by server, node status not updated

# POST /identify

### Description:
Allows nodes to identify and register a UUID and node version with a server

### Arguments:
A node metadata struct populated with the node's version and UUID, encoded in json


### Example:

```
http://host:8080/v0/identify?uuid=a468d5d0-56b8-4b0d-be2f-08b7d612b055&version=0.0.4
```

### Returns:
Nothing

### Status:
- 200 OK: Returns nothing, node UUID is now registered on this server
- 500 Internal Server Error: Error while processing identify request or registering this node
