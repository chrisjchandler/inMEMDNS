SpecialDNS: a Go based Authoratative DNS resolver that serves records stale until they are explicitly flushed

# DNS Resolver with Record Management

This is a DNS resolver implementation that allows you to manage DNS records and provides functionality to publish, modify, delete, and flush records. Additionally, it supports importing zone files.

## Record Management

### 1. Publishing Records

To publish a DNS record, you can use an API endpoint to add the record to the DNS resolver. The API endpoint should accept the necessary parameters, such as domain, record type, and record data. Here's an example of how to publish an "example.com A" record using `curl`:

```shell
curl -X POST -d '{"domain": "example.com", "type": "A", "data": "192.168.1.100"}' http://localhost:8080/dns/records

Adjust the endpoint URL and request body according to your API design and implementation.

2. Modifying Records
To modify an existing DNS record, you can use an API endpoint that allows updating the record data. The endpoint should accept the record identifier or any other parameters required to identify the record. Here's an example of how to modify the "example.com A" record using curl:


curl -X PUT -d '{"data": "192.168.1.200"}' http://localhost:8080/dns/records/example.com/A

Replace the endpoint URL and request body with the appropriate values based on your API design and implementation.

3. Flushing Stale Records
To flush all stale records, you can use a console command. In the DNS resolver console, enter the command flushstale. This command will clear all the records marked as stale.

4. Flushing Specific Records
To flush a specific record, you can use a console command followed by the record information. In the DNS resolver console, enter the command flushrecord and provide the record in the format "domain type". For example, to flush the "example.com A" record, enter:

flushrecord example.com A

5. Deleting Records
To delete a specific DNS record, you can use an API endpoint that accepts the record identifier or any other parameters required to identify the record. For example, to delete the "example.com A" record using curl:

curl -X DELETE http://localhost:8080/dns/records/example.com/A

Replace the endpoint URL with the appropriate value based on your API design and implementation.

6. Importing Zone Files
To import zone files, you can use a function in the DNS resolver code that reads and processes the zone file data. Implement a mechanism to parse the zone file format and add the records to the DNS resolver's data store. You can trigger this import process either programmatically or through a console command.

DNS Record Storage
In this implementation, DNS records are stored in two main data structures:

zoneData: Stores the active and valid DNS records grouped by domain. These records are served in response to DNS queries.

staleData: Stores the stale DNS records that have been marked for deletion but are still served as authoritative results until explicitly flushed. These records can be flushed using console commands or API endpoints.


