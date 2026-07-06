# Características de Catsplash 🐱

Este documento contiene una tabla detallada de las características actuales de **Catsplash**, comparadas con capacidades estándar de portales como *NoDogSplash*, marcando cuáles están implementadas y cuáles están pendientes para futuras versiones.

| Categoría | Característica | Estado | Descripción |
| :--- | :--- | :---: | :--- |
| **Intercepción y Redirección** | Redirección HTTP (Puerto 80) | ✅ | Intercepta peticiones HTTP y redirige al portal. |
| | Compatibilidad CNA (iOS/Android) | ✅ | Detecta y abre el mini-navegador nativo del sistema operativo. |
| | Reglas de Bypass por MAC/IP | ✅ | Evita la redirección para clientes ya autenticados. |
| | Redirección HTTPS Segura | ⚠️ *Solo Administrados* | Intercepción en puerto 443. Requiere instalar una CA local (generada con `mkcert`) en el dispositivo cliente. |
| **Gestión de Sesiones** | Persistencia en Base de Datos | ✅ | Almacena clientes y sesiones en base de datos SQLite. |
| | Expiración por Tiempo Absoluto | ✅ | Desconecta al cliente tras transcurrir el tiempo máximo de sesión. |
| | Expiración por Inactividad | ✅ | Libera la sesión si el cliente no genera tráfico en un periodo. |
| | Limpieza en Cierre (*Fail-Secure*) | ✅ | Restaura las reglas de `iptables` y bloquea accesos al apagar la app. |
| | **Aumentar Tiempo de Sesión Manualmente** | ❌ | Extender el tiempo de un cliente sin requerir que vuelva a autenticarse. |
| **Administración y Control** | Panel de Administración Web (Dashboard) | ❌ | Interfaz web para administradores donde ver estadísticas y gestionar clientes. |
| | **Ver Clientes Conectados en Tiempo Real** | ⚠️ *Parcial* | Solo visible directamente haciendo consultas SQL a la base de datos o mediante el script `setup.sh status`. |
| | **Bloquear/Desconectar Cliente Manualmente** | ❌ | Expulsar o banear a un cliente (por MAC/IP) desde la administración. |
| | API REST para Integración | ❌ | API para integrar Catsplash con sistemas de terceros. |
| **Autenticación** | Pantalla de Bienvenida (ToS/Click-to-Connect) | ✅ | El usuario acepta términos y condiciones para navegar. |
| | Sistema de Vales (Vouchers) | ❌ | Generación de códigos únicos temporales para dar acceso. |
| | Autenticación de Usuarios (User/Pass) | ❌ | Registro e inicio de sesión de cuentas de usuario locales. |
| | Integración con RADIUS (WPA Enterprise) | ❌ | Validación de usuarios contra servidores externos de autenticación. |
| **Control de Tráfico y QoS** | Limitación de Ancho de Banda (Rate Limiting) | ❌ | Controlar la velocidad de bajada y subida por usuario (usando `tc` de Linux). |
| | Cuotas de Consumo de Datos (MB/GB) | ❌ | Limitar la cantidad total de datos que un cliente puede transferir. |

---

## Inspiración en NoDogSplash y openNDS (`ndsctl`)

**NoDogSplash** (y su sucesor moderno **openNDS**) es el referente directo para **Catsplash**. Su arquitectura se basa en un demonio en segundo plano que escucha peticiones y se administra a través de una utilidad de consola llamada `ndsctl`.

### ¿Cómo funciona `ndsctl` en NoDogSplash?
`ndsctl` se comunica con el demonio del portal cautivo mediante un Socket de Dominio Unix (`/tmp/ndsctl.sock` o similar). Permite a los administradores realizar acciones en caliente sin reiniciar la aplicación:
*   `ndsctl status`: Resumen del sistema (uptime, clientes conectados totales, estado del firewall).
*   `ndsctl clients`: Lista legible con los datos de cada cliente (MAC, IP, volumen de datos consumidos, tokens).
*   `ndsctl json`: Devuelve la misma lista pero formateada en JSON, ideal para scripts de automatización.
*   `ndsctl auth <MAC>`: Fuerza la autenticación manual de un dispositivo para darle acceso libre.
*   `ndsctl deauth <MAC>`: Expulsa/desconecta al cliente inmediatamente de la red.
*   `ndsctl block <MAC>`: Coloca un dispositivo en la lista negra para evitar que pueda siquiera ver el portal.

---

## Nota sobre la Redirección HTTPS y `mkcert`

Interceptar tráfico HTTPS (puerto 443) para mostrar el portal cautivo presenta un desafío de seguridad inherente debido al diseño de SSL/TLS (diseñado precisamente para evitar ataques de tipo Man-in-the-Middle).

### ¿Es posible usar `mkcert` para solucionar esto?
**Sí, pero con limitaciones importantes según el tipo de dispositivo:**

* **Dispositivos Administrados (Entornos Corporativos/Escuelas):**
  * **Es viable.** Si utilizas `mkcert` para generar una Autoridad de Certificación (CA) raíz local y la instalas (mediante políticas de grupo de Windows AD, MDM o scripts) en el almacén de certificados de confianza de los dispositivos clientes, estos confiarán en los certificados que genere el portal para portales seguros, eliminando las advertencias de seguridad.
* **Dispositivos de Invitados (Redes Públicas/Móviles Personales):**
  * **No es viable en la práctica.** La CA raíz de `mkcert` es privada. Si un dispositivo de un invitado no la tiene instalada, al intentar interceptar HTTPS verá una advertencia de seguridad crítica (bloqueo HSTS). Instalar una CA externa de forma manual en smartphones es complejo y representa un riesgo grave de seguridad para el usuario.

### Enfoque Estándar de la Industria:
Para redes con usuarios invitados, se evita interceptar el puerto 443. En su lugar:
1. Se confía en el **Asistente de Conectividad Nativo (CNA)** del sistema operativo del cliente (iOS/Android/Windows), el cual realiza pruebas en HTTP (`http://connectivitycheck.gstatic.com/generate_204`) al conectarse a la red WiFi y levanta automáticamente el portal web.
2. Se instruye al usuario a ingresar a una página sin SSL (por ejemplo, `http://neverssl.com`) si el portal no se abre automáticamente.

---

## Próximos Pasos Recomendados para Catsplash

Para convertir a **Catsplash** en un portal cautivo completo de producción al estilo de *NoDogSplash*, se sugieren los siguientes bloques de desarrollo:

1. **Creación de `catsctl` (Controlador CLI):**
   * Implementar en Go una pequeña utilidad de línea de comandos (`catsctl`) que lea/escriba en el mismo archivo de base de datos SQLite (o use un socket interno si se requiere tiempo real estricto).
   * Comandos propuestos:
     * `catsctl status`: Mostrar estado de la base de datos y estadísticas rápidas.
     * `catsctl list`: Listar clientes con su IP, MAC y tiempo restante de sesión.
     * `catsctl auth <MAC> [duración]`: Autenticar o extender manualmente la sesión de una MAC.
     * `catsctl kick <MAC>`: Revocar acceso de red y aplicar reglas de bloqueo.

2. **Panel Web de Administración:**
   * Crear una sección web privada (`/admin`) protegida por contraseña dentro del propio binario web de Catsplash, para que el administrador pueda ejecutar estas acciones visualmente desde su móvil.

3. **Limitación de Ancho de Banda (QoS):**
   * Integrar la herramienta `tc` (Traffic Control) de Linux en el paquete `firewall` para aplicar colas de velocidad (ej. 2 Mbps de bajada, 512 Kbps de subida) a la IP de cada cliente que sea autenticado.
