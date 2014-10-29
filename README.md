mysql-jsondump
==============

Simple, fast MySQL table to JSON dumper

Usage
-----

mysql-jsondump HOST PORT USER PASSWORD DATABASE QUERY > output.json


Limitation
----------

Currently, mysql-jsondump dumps all values as strings.
Any normal RDB automatically converts string to any types, it may not be a problem.
