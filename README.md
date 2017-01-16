sqldump
=======

Simple, fast RDB table dumper


Usage
-----

SQLDUMP_PASSWORD=PASSWORD sqldump [--tsv] [--gzip] HOST PORT USER DATABASE QUERY > output_file


Limitation
----------

sqldump supports only MySQL still now.

sqldump dumps all field values as strings.
Many RDB automatically converts string to any types, it may not be a problem.


License
-------

MIT license.


Copyright
---------

Copyright (c) 2014-2017 Minero Aoki
