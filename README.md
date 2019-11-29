# svc1: Loc 36 Service 1




__service ID:__           1

__service Name:__         State recorder

__service Description:__  Service 1 is responsible for recording the state of a sensor, in the DBMS. Valid sensor reports (EMF state information) are recorded, in the DBMS.




### Using the service (Service 1)

Service 1 is expected to be consumed by sensors or sensor proxies. Service request can be made via only HTTPS.


#### To request a service:

	send an HTTPS POST request to the service, with the request path being: /report

	the request is expected to include the following field data:

		state:      Should be the state of the sensor.
		sensor:     Should be the id of the sensor.
		sensorPass: Should be the pass of the sensor.
		serviceId:  Should be the id of the service the client thinks its interacting with.
		serviceVer: Should be the service version the client needs.


#### Response:

	Service 1 should ordinarily return HTTP response code 200, and if any other code is returned. A fatal error should be assumed to have happened.

	If the HTTP response code is 200, a JSON data would also be provided:

		{
			response: "{x}",
			responseCode: "{y}"
		}

	This JSON data provides the response of the service request.

	Possible response codes are:

		a: State updated successfully! 
		b: An error occured.
		c: Incomplete request data.
		d: Service requested from the wrong service.
		e: Unsupported service version.
		f: State provided seems invalid.
		g: Unknown sensor.
		h: Incorrect sensor password.




## Building and deployment

Building service 1 is as simple as downloading a release, then running:

	go build

in the directory of its source code.


#### Deployment

- Put the executable of service 1 in the machine it is expected to run on;
- Copy files 'httpConf.yml' and 'conf.yml' to the executable's directory. These files can be found in the directory of the source code.
- Modify the files, as apprpriate.
