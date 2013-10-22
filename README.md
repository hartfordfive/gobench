<h2>gobench</h2>
=======

A simple HTTP stress testing tool <br/><br/>


Sample Usage Example:<br/>
-----------------------<br/>

go run gobench -u "http://www.yourdomain.com" -c 5 -m 25 -cf cookies.txt -ul useragentlist.txt<br/>

or with compiled binary:<br/>

./gobench -u "http://www.yourdomain.com" -c 5 -m 25 -cf cookies.txt -ul useragentlist.txt<br/><br/>


Mandatory Parameter Details:<br/>
------------------------<br/>
`-u` : The url to test**<br/>
`-m` : The total number of requests to send<br/>

Optional Parameter Details:<br/>
------------------------<br/>
`-c` : The number of requests to run concurently <br/>
`-p` : The number of processors to use <br/>
`-tw` : The number of milliseconds to wait between requests<br/>
`-pd` : The POST data file to use for the requets<br/>
`-l` : The file containing the list of URLs to request.  When specified, the `-u` parameter is ignored.**<br/>
`-cf` : The file containing the cookie data<br/>
`-ul` : The file contianing the list of user-agents to use for requests randomly. (One UA per line)<br/>


The following sample files have been included in the configs directory for the `-pd`, `-l`, `-cf` and `-ul` flags:<br/>

- cookies.txt<br/>
- postdata.txt<br/>
- ua.txt<br/>
- urls.txt<br/>