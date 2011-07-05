# pyaudience

This is a very simple interface stub for python to talk to a running
conductor.

submit_job submits a job.  It does no argument checking (it probably
should).

get_status gets data for a job.  If successful, it returns the unmarshall'd
json result.

Both methods throw exceptions if anything goes wrong.  There is a
ServerError exception type which will be raised if the server
complains about anything.

For more information about this API, please see doc/audience_api.txt
