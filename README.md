### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `

### Solucion Ejercicio N°3:

Como la consigna dice que netcat no puede ser instalado en la maquina host y que no se pueden exponer puertos del servidor, necesariamente hay que levantar un contenedor que se conecte a la red interna de los coontenedores `server` y `client`, que sea cpaaz de ejecutar netcat para interactuar con el server. 

Para verificar el nombre de la network se corrio el comando `docker network ls`:

![docker network ls](imgs/docker_network.png)

Se puede ver que la red se llama `tp0_testing_net` y a esta misma red hay que conectar al contenedor que va a ser capaz de correr netcat para probar el server. Una vez que los contenedores estan conectados a la misma red, pueden comunicarse utilizando la IP del contendor o los nombres.

Vamos a correr una imagen de [`busybox`](https://hub.docker.com/_/busybox) para crear un contenedor temporal (se corre con el tag `--rm` para que sea eliminado al terminar de ejecutar) a partir de esta imagen que ya tiene instalado netcat listo para ser usado. Se corre utilizando el siguiente comando:
```sh
docker run --rn --network="$NETWORK" busybox sh -c "echo $MESSAGE | nc server 12345"
```

Donde `--network` le dice que se conecte a esa network donde el contenedor del server esta conectado asi se puede comunicar. Ejecutando `docker run --help` se puede verificar el uso de `--network`:
--network network Connect a container to a network

Con `sh -c` se invoca a la terminal y se le indica que ejecute el comando que viene despues de `-c`.
Utilizando netcat se envia el mensaje al server (nombre del contenedor) y se guarda la respuesta.

Luego se verifica la respuesta obtenida desde el server y se imprime el mensaje correspondiente.