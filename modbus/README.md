# Mainflux Modbus Adapter

The Mainflux Modbus Adapter service is responsible for reading and writing data to Modbus sensors using various protocols such as TCP and RTU/ASCII. It serves as an interface between Mainflux and Modbus devices, allowing you to easily integrate Modbus devices into your IoT ecosystem.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                           | Description                              | Default                        |
| ---------------------------------- | ---------------------------------------- | ------------------------------ |
| MF_MODBUS_ADAPTER_LOG_LEVEL        | Service log level                        | info                           |
| MF_JAEGER_URL                      | Jaeger server URL                        | http://jaeger:14268/api/traces |
| MF_MODBUS_ADAPTER_INSTANCE_ID      | Modbus adapter instance ID               |                                |
| MF_MODBUS_ADAPTER_RPC_HOST        | Modbus service HTTP host                 |                                |
| MF_MODBUS_ADAPTER_RPC_PORT        | Modbus service HTTP port                 | 8855                           |

## Deployment

Check the [`modbus-adapter`](https://github.com/mainflux/edge/blob/master/docker/modbus/docker-compose.yml#L6) service section in
docker-compose to see how service is deployed.

Running this service outside of container requires working instance of the message broker service.
To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/mainflux/mainflux

cd mainflux

# compile the binary
make modbus

# copy binary to bin
make install

# set the environment variables and run the service
MF_MODBUS_ADAPTER_LOG_LEVEL=[Service log level] \
MF_MODBUS_ADAPTER_RPC_HOST=[Message broker instance URL] \
MF_MODBUS_ADAPTER_RPC_PORT=[Message broker instance URL] \
MF_JAEGER_URL=[Jaeger server URL] \
MF_MODBUS_ADAPTER_INSTANCE_ID=[CoAP adapter instance ID] \
$GOBIN/mainflux-modbus
```

## Usage

The Mainflux Modbus Adapter service interacts with Modbus sensors through an RPC interface to perform read and write operations.

## Configuration
Before using the service an RPC call needs to be made to the Configure method passing either RTU or TCP configuration as shown in the example.

```json
{
  "address": "/dev/ttyS0",
  "baud_rate": 9600,
  "config": {},
  "data_bits": 8,
  "idle_timeout": "5m",
  "parity": "even",
  "rs485": {},
  "slave_id": 1,
  "stop_bits": 1,
  "timeout": "10s",
  "sampling_frequency": "1s"
}
```

```json
{
  "address": "localhost:1502",
  "idle_time": "15m",
  "slave_id": 2,
  "timeout": "5s",
  "sampling_frequency": "30s"
}
```

```go
client, err := rpc.Dial("tcp", "localhost:8855")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var id int
	// configure
	config := modbus.TCPHandlerOptions{
		Address: "localhost:1502",
	}

	err = client.Call("Adapter.ConfigureTCP", config, &id)
	if err != nil {
		fmt.Println("Configure Error:", err)
	} else {
		fmt.Println("Configure Response:", id)
	}
```

### Reading Values

To start reading values, you need to perform an RPC call to the read method.

```go
configRead := modbus.RWOptions{
		Address:   100,
		Quantity:  1,
		DataPoint: modbus.HoldingRegister,
		ID:        id,
	}

	data := make([]byte, 1)

	// Read
	err = client.Call("Adapter.Read", configRead, &data)
	if err != nil {
		fmt.Println("Read Error:", err)
	} else {
		fmt.Println("Read Data:", hex.EncodeToString(data))
	}
```


The supported data points include:

- coil
- h_register
- i_register
- register
- discrete
- fifo


### Writing Values

To start writing values, you need to perform an RPC call to the write method.

```go
configWrite := modbus.RWOptions{
		Address:   100,
		Quantity:  1,
		Value:     modbus.ValueWrapper{Data: uint16(1)},
		DataPoint: modbus.Register,
		ID:        id,
	}

	// Write
	err = client.Call("Adapter.Write", configWrite, &data)
	if err != nil {
		fmt.Println("Write Error:", err)
	} else {
		fmt.Println("Write Response:", hex.EncodeToString(data))
	}
```

The value field can be either `uint16` or `[]byte`.

### Notes
More examples in the `example` dir.
Some simulators are available to get you started testing:
- https://github.com/TechplexEngineer/modbus-sim
