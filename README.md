### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).

### Solucion Ejercicio N°2:

En el archivo docker-compose-dev.yaml tanto para el server como para el cliente se agrega un volumen que monta los archivos de configuracion respectivamente en cada uno de los contenedores, haciendo que los cambios que se hagan en los archivos locales se reflejen automaticamente en los contenedores sin necesidad de reconstruir las imagenes. Esto permite persistir los archivos de configuracion. 

Se agrega al .dockerignore los archivos de configuracion ya que al crear un volumen en el container no es necesario copiarlos al crear las imagenes.

Ya no es necesario hacer rebuild de las imagenes cada vez que se hace un cambio en los archivos de configuracion. Los cambios se van a ver tambien en los contenedores.