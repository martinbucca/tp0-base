### Ejercicio N°8:
Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).


### Solucion Ejercicio N°8:


En los ejercicios anteriores, el servidor procesaba todos los mensajes de un cliente de manera **secuencial** y recien despues aceptaba a otro. Es por eso que cuando un cliente terminaba de enviar las apuestas, se tenia que desconectar y por cada vez que quisiera pedir los ganadores tenia que abrir una nueva conexion. De esta manera, el servidor podia procesar a otros clientes.

Con las modificaciones de este ejercicio, ahora el servidor puede procesar varios clientes en paralelo, permitiendo que las agencias interactuen con la Central de Loteria en simultaneo.


- Por cada cliente se lanza un thread
  ```python
  while self._is_currently_running:
    try:
        agency_client_sock = self.__accept_new_connection()
        t = threading.Thread(target=self.__handle_client_connection, args=(agency_client_sock,))
        t.start()
  ```

- Van a haber ciertos recursos que se van a tener que compartir entre los threads. Para garantizar propiedades Safety (exclusion mutua y ausencia de deadlocks), se usa un Lock que provee `threading`. Los recursos compartidos son:
  - `_agencies_finished`: Los threads tienen que poder actualizar esta variable
  - `_number_of_agencies`: Los threads usan esta variable para saber si ya se terminaron de procesar todos los clientes
  - `winners_are_ready`: Los threads tienen que poder acceder a esta variable
  - `store_bets`: Esta funcion accede y escribe un archivo, por lo que es importante que no ocurra que dos procesos quieran cambiarlo al mismo tiempo para evitar inconsistencias
  
- Para sincronizar los diferentes threads y que todos sepan que se termino de procesar a todos los clientes se usa un `Event`:
  ```python 
  self.winners_are_ready = threading.Event()
  ```
- Cuando todos los clientes terminan de enviar apuestas se setea el evento
   ```python 
  if self._agencies_finished == self._number_of_agencies:
    logging.info("action: sorteo | result: success")
    self.winners_are_ready.set()
  ```
- Cuando un cliente pide por los ganadores se asegura que todos los clientes esten listos haciendo un `wait` al evento. En caso de que se haya seteado va a avaznar y sino va a quedar esperando.
   ```python 
  if self._agencies_finished == self._number_of_agencies:
    self.winners_are_ready.wait():
      ...
  ```

- En el metodo `shutdown` se cierra el socket del servidor y se le hace `join()` a todos los threads para asegurar que no quede ninguno colgado.


### Estimacion del MaxBatchAmount Default

- Los paquetes que se envien no pueden superar los 8kB.
- Teniendo en cuenta que los mensajes que envian apuestas en batchs estan compuestos de la siguiente manera:
 - 2 Bytes fijos para el Id del mensaje
 - 2 Bytes fijos para el largo del payload
 - Payload es texto utf-8 separando campos por el caracter `"&"`:
   - Id Cliente ~ 10 Bytes
   - Id Chunk ~ 10 Bytes
   - Apuestas, donde cada apuesta es texto utf-8 separando campos por el caracter `"|"`:
     - Nombre ~ 20 Bytes
     - Apellido ~ 25 Bytes
     - Documento ~ 15 Bytes
     - Fecha de Naciemiento ~ 10 Bytes
     - Numero de Apuesta ~ 10 Bytes

--> El tamaño maximo aproximado de cada apuesta es:
```
20 B (Nombre) + 1 B (|) + 25 B (Apellido) + 1 B (|) + 15 B (Documento) + 1 B (|) + 10 B (Fecha de Nacimiento) + 1 B (|) + 10 B (Numero de Apuesta) + 2 B (& separador inicial y/o final)= 86 Bytes
```

--> Tamaño maximo aproximado para el resto del payload:
```
10 B (Id Cliente) + 1 B (&) + 10 B (Id chunk) + 1 B (&) = 22 Bytes
```
--> Tamaño maximo aproximado de cada paquete entonces va a ser:
```
4 Bytes (header) + 22 Bytes (inicio del payload) + 86 Bytes (cada apuesta) * N_APUESTAS
```
- Como esto tiene que ser menor a 8kB, queda:

```
86 * N_APUESTAS < 8192 - 4 - 22
N_APUESTAS < 8166 / 86
N_APUESTAS < 94.95
N_APUESTAS < 94
```



