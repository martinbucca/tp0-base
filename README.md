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
  - La funcion `handle_sigterm` ejecuta `shutdown_server` que setea la condicion `is_currently_running` a False y cierra el socket del Server. 
  - el `accept()` del socket es bloqueante pero tiene un timeout de 1s para que en caso de no recibir conexion entrante en ese tiempo, lanze un error y se verifique la condicion `is_currently_running` del loop while. 
  - Tambien cuando se cierre el socket del server, al intentar hacer `accept()` va a saltar un error que se atrapa en el main loop y al verificarse la variable `is_currently_running`, como va a ser False, se va a salir del while loop


- Client:
  - Al inicializar el cliente se setea la variable `is_currently_running` = `True`. Esta condicion 
  va a servir para chequear en el main loop del cliente.
  - Se configura un handler para manejar la señal SIGTERM (`setupSigtermHandler`). Se crea un canal (`sigChannel`) con capacidad para un 1 elemento que puede recibir señales del sistema y en caso de recibir una señal SIGTERM se le va a notificar. Se lanza una gorutine que ejecuta la funcion `handleSigterm` y queda bloqueada esperando que llegue una señal por el canal.
  - Cuando recibe una señal setea la variable  `is_currently_running` = `False` y cierra la conexion. En el loop principal del cliente, si se estaba intentando leer algo va a lanzar un error porque la conexion fue cerrada y se va a loggear el error y retornar. Si no, la condicion del loop va a detectar que `is_currently_running` = `False` y no va seguir ejecutando el loop. 

### Tests


![Tests Ejercicio 4](imgs/tests-ej4.png)