# do-serverless-ipinfo

Dos funciones HTTP en Go para DigitalOcean Functions. `ip/my` devuelve la primera dirección de la cabecera `X-Forwarded-For`. `ip/info` recibe una dirección IPv4 o IPv6 y devuelve el JSON de IPinfo.

## Desarrollo local

Desde la raíz del repositorio, instala Go 1.27.4 con mise y ejecuta las pruebas. Estas simulan las respuestas de IPinfo y se pueden ejecutar sin credenciales. Los comandos usan `GO111MODULE=off` porque el proyecto solo usa la biblioteca estándar y carece de un archivo `go.mod`.

```sh
mise trust
mise install
GO111MODULE=off mise exec -- go test ./...
GO111MODULE=off mise exec -- go vet ./...
```

## Despliegue

El archivo [project.yml](project.yml) configura ambas funciones con el runtime `go:1.25`. Con `doctl` autenticado, la extensión serverless instalada y un namespace de Functions conectado, despliega desde la raíz del repositorio.

```sh
doctl serverless deploy .
```

Por defecto, `ip/info` consulta IPinfo sin credenciales. Para usar tu cuenta, sustituye `environment: {}` **dentro de la función `info`** por el siguiente bloque antes de desplegar. Proporciona el valor de `IPINFO_ACCESS_TOKEN` al despliegue mediante un archivo de entorno privado. [Configuración de DigitalOcean Functions](https://docs.digitalocean.com/products/functions/reference/project-configuration/)

```yaml
environment:
  IPINFO_ACCESS_TOKEN: "${IPINFO_ACCESS_TOKEN}"
```

## Uso

Obtén la dirección que el proxy comunica a la función. Una respuesta de ejemplo es `{"ip":"203.0.113.7"}`.

```sh
curl "$(doctl serverless function get ip/my --url)"
```

Consulta una dirección concreta mediante el parámetro `ip`.

```sh
curl --get --data-urlencode 'ip=8.8.8.8' "$(doctl serverless function get ip/info --url)"
```

Ambas funciones devuelven JSON. En caso de error, el campo `error` contiene un mensaje en español.

| Estado HTTP | Cuándo se devuelve |
| --- | --- |
| 400 | Falta la dirección o es inválida. En `ip/my` se comprueba la primera dirección reenviada. |
| 429 | IPinfo ha limitado la frecuencia de las consultas. |
| 502 | Falló la conexión con IPinfo, el proveedor devolvió otro error o la respuesta no es JSON válido. |
| 504 | La consulta a IPinfo superó el límite de cinco segundos. |

## Licencia

[MIT](LICENSE).
