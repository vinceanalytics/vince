/*
Package timeseries implements a Key/Value store backed, RoaringBitmaps based
storage for web analytics events. We use pebble as embedded key value store and
rely heavily on its value merge feature, so it is not possible to support different
underlying key value store.

Web analytics event comprises of the folllowing fundamental properties.
NOTE: The structure is taken from earlier version of vince, we no longer use protocol
buffers but the name and datatype stays the same except for bounce, which is now
represented as an int8.

	int64 timestamp = 1;
	int64 id = 2;
	optional bool bounce = 3;
	bool session = 4;
	bool view = 5;
	double duration = 6;
	string browser = 19;
	string browser_version = 20;
	string city = 26;
	string country = 23;
	string device = 18;
	string domain = 25;
	string entry_page = 9;
	string event = 7;
	string exit_page = 10;
	string host = 27;
	string os = 21;
	string os_version = 22;
	string page = 8;
	string referrer = 12;
	string region = 24;
	string source = 11;
	string utm_campaign = 15;
	string utm_content = 16;
	string utm_medium = 14;
	string utm_source = 13;
	string utm_term = 17;
	string tenant_id = 28;

This is the only data structure we need to store and query effiiciently. All string
prpperties are used for search and aggregation.

# Timeseries

All queries going through this package are time based. Computation of time ranges
and resolutions is handled by the internal/compute package.

We have six time resolutions that is used for search

  - Minute
  - Hour
  - Day
  - Week
  - month
  - Year

Time in unix millisecond truncated to above resolution is stored as part of keys
in a way that when querying similiar timestamp will load  similar blocks speeding
up  data retrieval. Details about timestamp encoding will be discusssed in the
Keys section.

# Keys

# A key is broken into the following components

[ byte(prefix) ][ uint64(shard) ][ uint64(timestamp) ][ byte(field) ]

prefix: encodes a unique global prefix assigned for timeseries data. This value
is subject to change, however it is the sole indicator that the key  holds time
series data.

shard: We store in 1 Million events partitions. Each event gets assigned a unique ID
that is auto incement of uint64 value. To get the assigned shard.

	shard = id / ( shard_width ) # shard_width = 1 << 20

field: we assign unique number to each property.

	Field_unknown           Field = 0
	Field_timestamp         Field = 1
	Field_id                Field = 2
	Field_bounce            Field = 3
	Field_duration          Field = 4
	Field_city              Field = 5
	Field_view              Field = 6
	Field_session           Field = 7
	Field_browser           Field = 8
	Field_browser_version   Field = 9
	Field_country           Field = 10
	Field_device            Field = 11
	Field_domain            Field = 12
	Field_entry_page        Field = 13
	Field_event             Field = 14
	Field_exit_page         Field = 15
	Field_host              Field = 16
	Field_os                Field = 17
	Field_os_version        Field = 18
	Field_page              Field = 19
	Field_referrer          Field = 20
	Field_source            Field = 21
	Field_utm_campaign      Field = 22
	Field_utm_content       Field = 23
	Field_utm_medium        Field = 24
	Field_utm_source        Field = 25
	Field_utm_term          Field = 26
	Field_subdivision1_code Field = 27
	Field_subdivision2_code Field = 28

shard and timestamp compoenents are encodded as binary.AppendUvarint. This scheme
ensures efficient time range queries. We can effficiently iterate on co located
data most of the times.
*/
package timeseries
