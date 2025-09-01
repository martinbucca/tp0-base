#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Uso: $0 <archivo_salida> <cantidad_clientes>"
    exit 1
fi

ARCHIVO_SALIDA=$1
CANT_CLIENTES=$2

echo "Nombre del archivo de salida: $ARCHIVO_SALIDA"
echo "Cantidad de clientes: $CANT_CLIENTES"
python3 mi-generador.py "$ARCHIVO_SALIDA" "$CANT_CLIENTES"
