[Godoc](http://godoc.org/github.com/getlantern/elevated)

elevated currently only works on OS X and Windows.

See the demo programs for an example of using elevated.

The OS X demo program uses the networksetup utility to adjust the MTU on
interface en0, which requires root permissions.

The Windows demo program (which only works on Windows 7+ right now) uses the
netsh utility to update firewall settings.  The setting that we're changing
requires admin privileges.

Both demo programs try their respective operations with and without elevation to
demonstrate that only the one with elevation works.