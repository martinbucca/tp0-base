### Ejercicio N.º 5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Solucion Ejercicio N°5:

### Protocolo

----------------------------------------------------------------------------------------------------
Largo_del_payload|ID_MENSAJE|ID_Agencia|Nombre|Apellido|Documento|Fecha_nacimiento|Numero

------4 Bytes----  -----------------------------   Variable ----------------------------------------

Los primeros 4 bytes van a ser para indicar el largo del payload. El payload va a contener toda la informacion sobre la apuesta y va a ser de tamaño variable.

1. El cliente (Agencia de quiniela) le manda un mensaje al servidor (Central de Loteria Nacional) en donde el ID del mensaje es `BET`.
2. El servidor lee los primeros 4 bytes y sabe exactamente cuantos bytes mas tiene que leer para el payload. Deserializa el payload y separa los campos que vienen en el orden indicado en la estructura  y separados por el caracter `|`
3. En caso de poder guardar correctamente la apuesta, el servidor le va a mandar un mensaje al cliente en donde los primeros 4 bytes son el largo del payload y un payload variable. Este protocolo define doss posibles mensajes para ese payload:
- "OK": En caso de que se pudo guardar la apuesta correctamente
- "ERROR": En caso de algun error en el servidor para procesar la apuesta
------------------------------------
Largo_del_payload|MENASAJE

------4 Bytes----  ----Variable------


### Tests


![Tests Ejercicio 5](imgs/tests-ej5.png)
