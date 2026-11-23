# do-serverless-ipinfo

Dos funciones HTTP en Go para DigitalOcean Functions. `ip/my` devuelve la primera dirección de la cabecera `X-Forwarded-For`, su familia y la información de IPinfo. `ip/info` recibe una dirección IPv4 o IPv6 y devuelve el JSON de IPinfo.

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

Por defecto, ambas funciones consultan IPinfo sin credenciales. Para usar tu cuenta, sustituye `environment: {}` **en el paquete `ip`** por el siguiente bloque antes de desplegar. Proporciona el valor de `IPINFO_ACCESS_TOKEN` al despliegue mediante un archivo de entorno privado. [Configuración de DigitalOcean Functions](https://docs.digitalocean.com/products/functions/reference/project-configuration/)

```yaml
environment:
  IPINFO_ACCESS_TOKEN: "${IPINFO_ACCESS_TOKEN}"
```

## Uso

Para ver tus dos direcciones de salida en un único JSON, ejecuta desde la raíz del repositorio (requiere Python 3 y curl):

```sh
./bin/mis-ips
```

El comando se conecta directamente a `ip.efrenfuentes.net` una vez por IPv4 y otra por IPv6. Cada clave, `ipv4` o `ipv6`, contiene `ip`, `family` e `info`. Si la consulta de una familia falla, su valor es `null` y el diagnóstico se escribe en stderr. El comando termina con código 1 si fallan ambas. Con un despliegue anterior de `ip/my`, consulta también `ip/info` para completar cada resultado.

Para consultar una sola conexión, llama a `ip/my` directamente. Si IPinfo falla, la función conserva la IP observada y devuelve HTTP 200 con `info: null` y un campo `error`.

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
| 429 | En `ip/info`, IPinfo ha limitado la frecuencia de las consultas. |
| 502 | En `ip/info`, falló la conexión con IPinfo, el proveedor devolvió otro error o la respuesta no es JSON válido. |
| 504 | En `ip/info`, la consulta a IPinfo superó el límite de cinco segundos. |

## Licencia

[MIT](LICENSE).
