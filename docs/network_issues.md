# Problemas de Conectividad y el Signo "!"

Es común que tras conectarse y autenticarse en un portal cautivo, el dispositivo siga mostrando un signo de admiración (`!`) en el icono de WiFi o indique "Conectado sin internet" por unos segundos. Aquí explicamos por qué sucede y cómo se soluciona.

## 1. El Signo "!" (Estado de Limbo)

Este icono no significa necesariamente que no tengas internet, sino que el sistema operativo **aún no ha verificado** la conexión tras el cambio de estado.

### ¿Cómo verifica el móvil si hay internet?
Los móviles Android e iOS intentan descargar un archivo pequeño (usualmente un HTTP 204) de servidores como `connectivitycheck.gstatic.com` o `captive.apple.com`.
*   **Antes de autenticar:** La petición es interceptada por `catsplash`. El móvil ve que la red está restringida.
*   **Justo después de autenticar:** El móvil puede tardar desde 5 hasta 60 segundos en volver a intentar esta comprobación. Durante ese intervalo, verás el signo `!`.

## 2. El Problema de la Redirección Automática

`catsplash` intenta redirigir a Google tras el éxito. Anteriormente, esto podía fallar si el firewall seguía redirigiendo el tráfico HTTP incluso después de la autenticación.

**Mejora implementada:** `catsplash` ahora añade automáticamente una regla de "Bypass" en la tabla NAT para cada cliente autenticado. Esto significa que tan pronto como el firewall te libera, las peticiones al puerto 80 ya no son interceptadas, permitiendo que el móvil verifique el internet sin obstáculos.

Sin embargo, el problema puede persistir brevemente por:
1.  **Caché DNS:** El móvil aún recuerda que "Google" era la IP del portal.
2.  **Cierre del CNA:** El "Asistente de Red" (la ventana que se abre sola) es un navegador limitado. A veces se cierra antes de procesar la redirección.

---

## 3. Soluciones Rápidas

Si después de darle a "Conectar" sigues viendo el `!` o no navegas:

1.  **Espera 10 segundos:** Dale tiempo al sistema para refrescar las reglas de red.
2.  **Intenta navegar manualmente:** Abre tu navegador real (Chrome, Safari, Firefox) y entra a una web sencilla (ej. `http://1.1.1.1` o `http://neverssl.com`). Esto fuerza al sistema a usar la nueva ruta.
3.  **Apaga y enciende el WiFi:** Esto obliga al móvil a re-evaluar la red desde cero. Al conectarse de nuevo, verá que ya tiene paso libre y el `!` desaparecerá inmediatamente.
4.  **Usa el botón "Finalizar":** En la página de éxito de `catsplash`, hemos añadido un botón manual por si la redirección automática falla.

---

## 4. Nota sobre HTTPS

El portal cautivo solo puede interceptar tráfico **HTTP (puerto 80)**. Si intentas entrar directamente a una web **HTTPS (puerto 443)** como Facebook o Gmail antes de autenticarte, el navegador mostrará un error de certificado o simplemente no cargará. Esto es normal y es una medida de seguridad de internet. **Siempre intenta entrar a una web HTTP para que el portal aparezca.**
