# Seguridad de la Red WiFi (WPA2-PSK)

Aunque el portal cautivo gestiona el acceso a Internet, es altamente recomendable proteger el tráfico inalámbrico mediante cifrado WPA2. Esto evita que atacantes cercanos puedan interceptar los datos que viajan entre los dispositivos y tu punto de acceso.

## 1. ¿Por qué usar contraseña en el WiFi?

*   **Cifrado del aire:** Sin contraseña, cualquier dato enviado (incluyendo credenciales si no usas HTTPS en todo) puede ser capturado por terceros.
*   **Control de Acceso de Capa 2:** Evita que usuarios no deseados siquiera se asocien a tu red, ahorrando recursos en `dnsmasq` y `hostapd`.
*   **Privacidad:** Asegura que la comunicación sea privada entre el cliente y el AP.

---

## 2. Configuración de hostapd con WPA2

Para activar la seguridad, modifica tu archivo `/etc/hostapd/hostapd.conf` añadiendo las siguientes líneas:

```conf
# --- Configuración Base ---
interface=wlx1cbfce41183a
driver=nl80211
ssid=Mi_Red_WiFi_Segura
hw_mode=g
channel=6

# --- Configuración de Seguridad (WPA2) ---
# 1 = WPA, 2 = IEEE 802.11i/RSN (WPA2)
wpa=2

# Contraseña de la red (mínimo 8 caracteres)
wpa_passphrase=TuContraseñaSegura123

# Algoritmos de gestión de claves
wpa_key_mgmt=WPA-PSK

# Protocolos de cifrado
# CCMP es el estándar para WPA2 (AES)
rsn_pairwise=CCMP

# --- Estabilidad ---
auth_algs=1
wmm_enabled=0
```

---

## 3. Aplicar los Cambios

Reinicia el servicio de `hostapd` para que la red comience a pedir contraseña:

```bash
sudo systemctl restart hostapd
```

---

## 4. Interacción con Catsplash

El uso de una contraseña de WiFi **no interfiere** con el portal cautivo. El flujo para el usuario será:
1.  Seleccionar la red WiFi.
2.  Introducir la contraseña de la red.
3.  Una vez conectado a la señal, el sistema operativo detectará el portal y abrirá la ventana de `catsplash`.
4.  El usuario deberá "Conectar" en el portal para tener acceso real a Internet.

Este modelo de "Doble Seguridad" es el estándar para redes corporativas y privadas que desean control total sobre sus usuarios.
