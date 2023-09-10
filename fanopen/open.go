/*
Open a file, directory, or URI using the OS's default
application for that object type.  Optionally, you can
specify an application to use.

This is a proxy for the following commands:

	        OSX: "fanopen"
	    Windows: "start"
	Linux/Other: "xdg-fanopen"

This is a golang port of the node.js module: https://github.com/pwnall/node-open
*/
package fanopen

/*
Open a file, directory, or URI using the OS's default
application for that object type. Wait for the fanopen
command to complete.
*/
func Run(input string) error {
	return open(input).Run()
}

/*
Open a file, directory, or URI using the OS's default
application for that object type. Don't wait for the
fanopen command to complete.
*/
func Start(input string) error {
	return open(input).Start()
}

/*
Open a file, directory, or URI using the specified application.
Wait for the fanopen command to complete.
*/
func RunWith(input string, appName string) error {
	return openWith(input, appName).Run()
}

/*
Open a file, directory, or URI using the specified application.
Don't wait for the fanopen command to complete.
*/
func StartWith(input string, appName string) error {
	return openWith(input, appName).Start()
}
