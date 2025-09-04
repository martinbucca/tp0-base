import sys

def parse_args():
    if len(sys.argv) != 3:
        print("Uso: python3 mi-generador.py <archivo_salida> <cantidad_clientes>")
        sys.exit(1)
    else:
        archivo_salida = sys.argv[1]
        cantidad_clientes = int(sys.argv[2])
        return archivo_salida, cantidad_clientes
    
def get_yaml_content(cantidad_clientes):
    content = """name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - NUMBER_OF_AGENCIES={cantidad_clientes}
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini

"""
    for i in range(1, cantidad_clientes + 1):
        content += f"""
  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
    networks:
      - testing_net
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-{i}.csv:/agency.csv
    depends_on:
      - server
"""
    content += """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""
    return content

def write_yaml_file(filename, content):
    with open(filename, 'w') as f:
        f.write(content)


def main():
    archivo_salida, cantidad_clientes = parse_args()
    yaml_content = get_yaml_content(cantidad_clientes)
    write_yaml_file(archivo_salida, yaml_content)
    print(f"Archivo {archivo_salida} generado con {cantidad_clientes} clientes.")

if __name__ == "__main__":
    main()