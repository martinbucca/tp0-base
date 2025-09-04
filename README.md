### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).

### Solucion Ejercicio N°2:

En el archivo `docker-compose-dev.yaml`, tanto para el servidor como para el cliente, se agrega un volumen que monta los archivos de configuración respectivos en cada uno de los contenedores. Esto permite que los cambios realizados en los archivos locales se reflejen automáticamente en los contenedores, sin necesidad de reconstruir las imágenes, y que la configuración persista por fuera de la imagen.

Además, se agregan los archivos de configuración al `.dockerignore`, ya que al utilizar volúmenes no es necesario copiarlos al crear las imágenes.

De esta manera, ya no es necesario reconstruir las imágenes cada vez que se modifica la configuración; los cambios se verán reflejados directamente en los contenedores.


```
volumes:
      - ./server/config.ini:/config.ini

volumes:
      - ./client/config.yaml:/config.yaml
```

### Tests

![Tests Ejercicio 2](imgs/tests-ej2.png)