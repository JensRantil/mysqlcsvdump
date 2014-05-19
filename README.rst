MySQL CSV dumper
================
.. image:: https://secure.travis-ci.org/JensRantil/mysqlcsvdump.png?branch=master
   :target: http://travis-ci.org/#!/JensRantil/mysqlcsvdump

This is a tiny program that can be used to _remotely_ dump one, or many, MySQL
tables to CSV format. The application was developed out of frustration that
``mysqldump`` still (in 2014!!!) can't dump to CSV format without storing the
dumped file on the database host.

Why CSV?
--------
`Comma-Separated Values`_ (CSV) is a very natural dataformat for relational
tables. It's readable by most applications that deals with data (yes, I'm
thinking about you, Excel) and it's a natural fit when working with data on the
command line (awk, sed, grep etc.).

.. _Comma-Separated Values: https://en.wikipedia.org/wiki/Comma-separated_values

Likewise, it can also `easily be parsed`_ by various scripting languages.

.. _easily be parsed: https://docs.python.org/2/library/csv.html

Why is dumping CSV files onto database instances a bad idea?
------------------------------------------------------------
TLDR; Because unnecessary fiddling on a database machine is generally a bad
idea. Hands off database machines if possible.

The longer answer:

First and foremost, I've seen many mistakes being made on database hosts.
Dumping to local disk will require you to log into the host, copy files,
possibly compress files and delete files. Many of these steps run the risk of
disturbing the work of the MySQL instance; IO could get worse or you run the
risk of running out of disk space. Heck, you even run the risk of deleting
files you did not intend to remove.

To avoid running out of disk space and not drain IO, you could of course mount
a separate file system for dumps - but I would rather not do that kind of work
on a database instance if not really necessary.

Also, dumping to local disk does not support compression. This utility
application does.

How do compile this?
--------------------
1. Install a Go compiler.

2. Set up a Go workspace: ``mkdir -p ~/src/mysqlcsvdump && export
   GOPATH=~/src/mysqlcsvdump``. See http://golang.org/doc/code.html for more
   information.

3. Get: ``go get github.com/JensRantil/mysqlcsvdump``.

3. Compile and install: ``go install github.com/JensRantil/mysqlcsvdump``.

4. Execute: ``~/src/mysqlcsvdump/bin/mysqlcsvdump -help``

Who developed this?
-------------------
I'm Jens Rantil. Have a look at `my blog`_ for more info on what I'm working
on.

.. _my blog: http://jensrantil.github.io/pages/about-jens.html
