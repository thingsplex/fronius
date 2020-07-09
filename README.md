# Futurehome Fronius Adapter

Adapter connects with Fronius inverter through local network and retrieves system data such ass current power production, total power production same day, same month, etc. 
Fronius inverter and Futurehome smarthub needs to be on the same network. 

Use `make deb-arm` to make package. 

Find the local IP-address of your Fronius inverter and save it in settings in the Fronius app in playgrounds. You can find the IP-address through the solar.web mobile app, or by scanning your network using tools such as Fing or similar. 

After saving the IP-address your Fronius inverter will appear in your device list within one minute.

***

The Fronius device will display current power production and total power production on current day. Currently it works as a `meter_elec` service, but will likely in the future be converted to `inverter` service.

***

## Services and interfaces
#### Service name
`meter_elec`
#### Interfaces
Type | Interface               | Value type | Description
-----|-------------------------|------------|------------------
out   | evt.meter_ext.report     | float_map       | map of current power production and total power production on same day