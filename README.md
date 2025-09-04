### Ejercicio N.º 4:

Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar
la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos,
sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear
mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando
`docker compose down`).

### Solucion Ejercicio N°4:

Tanto el servidor como el cliente necesitan manejar la señal SIGTERM para poder cerrar todos los recursos correctamente antes de que el thread principal muera.

- Server:
  - Al inicializar el servidor se setea la variable `is_currently_running` = `True`. Esta condicion va a servir para chequear en el main loop del server.
  - Se indica que al recibir la señal se llame a la funcinon `handle_sigterm`
    ```python
    signal.signal(signal.SIGTERM, self._handle_sigterm)
    ```
  - La funcion `handle_sigterm` ejecuta `shutdown_server` que setea la condicion `is_currently_running` a False y cierra el socket del Server.
    ```python
    def shutdown(self):
        try:
            self._is_currently_running = False
            self._server_socket.close()
            logging.info("action: shutdown | result: success | details: server socket closed")
        except Exception as e:
            logging.error(f"action: shutdown | result: fail | error: {e}")
    ```

  - El `accept()` del socket es bloqueante pero tiene un timeout de 1s para que en caso de no recibir conexion entrante en ese tiempo, lanze un error y se verifique la condicion `is_currently_running` del loop while. 
  - Tambien cuando se cierre el socket del server, al intentar hacer `accept()` va a saltar un error que se atrapa en el main loop y al verificarse la variable `is_currently_running`, como va a ser False, se va a salir del while loop
    ```python
    while self._is_currently_running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except (OSError, socket.timeout):
                if not self._is_currently_running:
                    break 
    ```
  - Para los sockets del cliente se espera que desde el cliente se cierre la conexion y al intentar hacer `recv()` o `send()` al socket del cliente lanze un error y salga del `recv()` bloqueante. Como hay una instruccion `finally` al finalizar la funcion ya sea por error o por flujo, se va a cerrar el socket del cliente
    ```python
    finally:
            client_sock.close()
    ```


- Client:
  - Al inicializar el cliente se setea la variable `is_currently_running` = `True`. Esta condicion 
  va a servir para chequear en el main loop del cliente.
  - Se configura un handler para manejar la señal SIGTERM (`setupSigtermHandler`). Se crea un canal (`sigChannel`) con capacidad para un 1 elemento que puede recibir señales del sistema y en caso de recibir una señal SIGTERM se le va a notificar. Se lanza una gorutine que ejecuta la funcion `handleSigterm` y queda bloqueada esperando que llegue una señal por el canal.
    ```go
    func setupSigtermHandler(c *Client) <-chan os.Signal {
        sigChannel := make(chan os.Signal, 1)
        signal.Notify(sigChannel, syscall.SIGTERM)
        go handleSigterm(c, sigChannel)
        return sigChannel
    }
    ```
  - Cuando recibe una señal setea la variable  `is_currently_running` = `False` y cierra la conexion. En el loop principal del cliente, si se estaba intentando leer algo va a lanzar un error porque la conexion fue cerrada y se va a loggear el error y retornar. Si no, la condicion del loop va a detectar que `is_currently_running` = `False` y no va seguir ejecutando el loop.
    ```go
    for msgID := 1; c.is_currently_running && msgID <= c.config.LoopAmount; msgID++ {
      ...
      msg, err := bufio.NewReader(c.conn).ReadString('\n')
      if c.conn != nil{
        c.conn.Close()
        c.conn = nil
      }
      ...
    }
    ``` 

### Tests


![Tests Ejercicio 4](imgs/tests-ej4.png)