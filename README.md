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
   ```python 
  if self._agencies_finished == self._number_of_agencies:
    logging.info("action: sorteo | result: success")
    self.winners_are_ready.set()
  ```
   ```python 
  if self._agencies_finished == self._number_of_agencies:
    if self.winners_are_ready.is_set():
      ...
  ```

- En el metodo `shutdown` se cierra el socket del servidor y se le hace `join()` a todos los threads para asegurar que no quede ninguno colgado.



