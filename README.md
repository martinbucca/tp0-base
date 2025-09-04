### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.


### Solucion Ejercicio N°6:


### Protocolo

### Client

El cliente puede enviarle los siguientes mensajes al servidor:
- SendBets

  ![sendBets message](imgs/sendBets.png)


  El mensaje esta compuesto por:
  - 2 Bytes para el ID del mensaje (fijos)
  - 2 Bytes para el largo del payload (fijos)
  - Payload cuyos bytes representan una cadena de texto (utf-8) que esta separada por el caracter <span style="color:red">"&" </span> y tiene los siguientes elementos:
    - Id Cliente (agencia)
    - Id Chunk
    - Apuestas, donde cada apuesta a su vez es una cadena de texto que separa los campos de la apuesta por el caracter <span style="color:blue">"|" </span>:
      - Nombre
      - Apellido
      - Documento
      - Fecha de Nacimiento
      - Numero de apuesta
    
    Ejemplo de un posible payload: 
    1<span style="color:red">&</span>43<span style="color:red">&</span>Santiago Lionel<span style="color:blue">|</span>Lorca<span style="color:blue">|</span>30904465<span style="color:blue">|</span>1999-03-17<span style="color:blue">|</span>2201<span style="color:red">&</span>Joaquin Sebastian<span style="color:blue">|</span>Rivera<span style="color:blue">|</span>21104770<span style="color:blue">|</span>1980-12-11<span style="color:blue">|</span>7737<span style="color:red">&</span>...

- sendFinsih:

  ![sendFinish message](imgs/sendFinish.png)

  El mensaje esta compuesto por:
  - 2 Bytes para el ID del mensaje (fijos)
  - 4 Bytes para el ID de la agencia (fijos)


### Server

El server puede enviarle los siguientes mensajes al Cliente:
  - sendOk
    
    ![sendOk](imgs/sendOk.png)

    El mensaje esta compuesto por:
      - 2 Bytes fijos para el ID del mensaje
      - 4 Bytes fijos para el chunk Id


  - sendFinishACK

    ![sendFinish](imgs/sendFinish.png)

    El mensaje es igual al que envia el cliente. Esta compuesto por:
      - 2 Bytes fijos para el ID del mensaje
      - 4 Bytes fijos para el ID de la agencia





Se tiene un CSVReader que lee un archivo csv linea por linea. 
Puede leer un chunk, recibiendo la cantiad de apuestas que quiere que tenga ese chunk como maximo y un id para ese chunk:

`func (r *CSVReader) ReadChunk(chunkId string, maxAmount int) (*BetsChunk, error)`
Lee hasta `maxAmount` lineas del archivo CSV y devuelve un struct que tiene las apuestas y el id del chunk:
```go
type BetsChunk struct {
	Bets []*Bet
	Id   string
}
```

1. Se envia el mensaje `sendBets` al servidor con el chunk
2. Se espera por el ack de ese mensaje
3. El servidor recibe el mensaje sendBets, almacena las apuestas y envia el ack para ese mensaje con sendOk que incluye el **chunk Id** del chunk almacenado.
4. Cuando se termina de leer todo el archivo el cliente envia el mensaje `SendFinish` con su ID
5. El servidor responde un ack con un mensaje igual `sendFinish` y el id de la agencia correspondiente.
6. El Cliente recibe el ack y finaliza.