A nivel conceptual, un portal cautivo no es un único programa, sino una **intersección de tres roles de red**: control de acceso físico/enlace, una pasarela de enrutamiento y un servidor de aplicaciones que interactúa con el usuario.

---

## Requerimientos Funcionales

Los requerimientos funcionales definen las acciones específicas que el sistema debe ser capaz de ejecutar para que el flujo del usuario se complete con éxito.

### 1. Intercepción y Aislamiento de Tráfico (Pre-Autenticación)

* **Restricción de tránsito por defecto:** El sistema debe denegar el acceso a la red externa (Internet) a cualquier nuevo dispositivo que se asocie a la interfaz de red local.
* **Permisión de servicios base locales:** El sistema debe permitir el libre tránsito de solicitudes de configuración automática de red (como la asignación de direcciones IP) y consultas de resolución de nombres locales antes de que el usuario se autentique. Si estos servicios se bloquean, el dispositivo cliente jamás podrá iniciar el flujo.

### 2. Redirección Forzada (Secuestro de Tráfico HTTP)

* **Detección de solicitudes de validación:** El sistema debe monitorear los intentos del cliente de comunicarse con el exterior a través de protocolos web estándar no cifrados.
* **Desvío de peticiones:** Al detectar tráfico web saliente de un usuario no autenticado, el sistema debe responder interceptando la solicitud original y devolviendo una instrucción que obligue al navegador del cliente a cargar una interfaz web local específica (el portal de bienvenida).

### 3. Gestión de Estado y Autenticación

* **Ciclo de vida del cliente:** El sistema debe mantener un registro en tiempo real de los dispositivos conectados utilizando identificadores unívocos de hardware (como las direcciones MAC) y sus direcciones lógicas (IP).
* **Cambio de estado dinámico:** Tras una acción afirmativa del usuario en la interfaz web (aceptar términos, iniciar sesión, ver un anuncio), el sistema debe cambiar el estado del cliente de *Pre-autenticado* a *Autenticado*.
* **Modificación de privilegios en caliente:** Al cambiar el estado, el sistema debe levantar inmediatamente las restricciones sobre el identificador de hardware del cliente, permitiendo que sus paquetes transiten libremente hacia la red externa.

### 4. Control de Sesión y Desconexión

* **Expiración por inactividad:** El sistema debe evaluar el tráfico del cliente y revocar su acceso si no se detecta actividad durante un periodo de tiempo predefinido.
* **Expiración por tiempo absoluto:** El sistema debe finalizar la sesión y devolver al usuario al estado de aislamiento una vez cumplido el tiempo máximo de navegación permitido, obligándolo a interactuar nuevamente con el portal.

---

## Requerimientos No Funcionales

Los requerimientos no funcionales definen las propiedades, restricciones de rendimiento, seguridad y experiencia que el sistema debe garantizar para operar correctamente.

### 1. Usabilidad y Compatibilidad Universal (Estándares de Interoperabilidad)

* **Compatibilidad con Asistentes Nativos de Conectividad (CNA):** El portal de bienvenida debe ser lo suficientemente liviano y estándar para renderizarse correctamente en los mini-navegadores integrados que los sistemas operativos móviles (Android, iOS, Windows) despliegan automáticamente al detectar una red restringida.
* **Independencia de configuración en el cliente:** El sistema debe operar sin requerir que el usuario instale aplicaciones adicionales, certificados raíz, ni que modifique manualmente los parámetros de red de su dispositivo.

### 2. Rendimiento y Latencia

* **Bajo impacto en servicios críticos iniciales:** El proceso de inspección y redirección de paquetes no debe degradar el tiempo de respuesta de la asignación de direccionamiento lógico ni de la resolución de nombres locales. La latencia introducida en el análisis de paquetes debe ser imperceptible para evitar que el cliente asuma que la red está caída.
* **Concurrencia de conexiones simultáneas:** El componente encargado de mantener la tabla de estados de los clientes debe ser capaz de procesar múltiples solicitudes de cambio de estado en paralelo sin bloquear el tráfico de los usuarios que ya se encuentran navegando.

### 3. Seguridad y Aislamiento de Clientes

* **Aislamiento de tráfico local (Seguridad en Capa 2):** El sistema debe garantizar que los dispositivos en estado de pre-autenticación no puedan comunicarse entre sí ni husmear el tráfico de otros usuarios de la red local.
* **Persistencia ante fallos estructurales:** Si el servicio que despliega la interfaz web se detiene o colapsa, el sistema de control de tráfico debe cerrarse de manera segura (*fail-secure*), manteniendo bloqueado el acceso a la red externa en lugar de dejarla abierta por accidente.

### 4. Eficiencia de Recursos (Escalabilidad)

* **Minimalismo en almacenamiento local:** Dado que el sistema suele operar en la frontera de la red (puertas de enlace o routers con recursos limitados), la interfaz web y el motor de control deben requerir una huella mínima de memoria volátil y almacenamiento masivo.

---

En resumen, los mínimos absolutos para que este sistema exista son: un mecanismo que **aísle**, un mecanismo que **engañe al navegador** para mostrar la pantalla de bienvenida, un mecanismo que **cambie el permiso** al hacer clic, y un mecanismo que **limpie las reglas** cuando el tiempo expire.
