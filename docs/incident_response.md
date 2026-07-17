# Plan de Respuesta a Incidentes — Catsplash

> Última actualización: 2026-07-16
> Referencia: LOPDP del Ecuador, Art. 50 (Medidas de seguridad)
> Vinculado a: `docs/dpia.md` §4, `docs/privacy_policy.md` §8

---

## 1. Alcance

Este documento define el procedimiento para detectar, contener, evaluar y notificar incidentes de seguridad que afecten datos personales tratados por Catsplash.

Un **incidente de seguridad** es cualquier evento que comprometa la confidencialidad, integridad o disponibilidad de datos personales, incluyendo:

- Acceso no autorizado a la base de datos SQLite
- Fuga de direcciones MAC o direcciones IP de usuarios
- Manipulación de reglas de firewall
- Compromiso de credenciales de administración
- Ataque de denegación de servicio contra el portal cautivo
- Brecha que exponga datos a terceros no autorizados

---

## 2. Clasificación de incidentes

| Nivel | Descripción | Ejemplos | Tiempo de respuesta |
|-------|-------------|----------|---------------------|
| **Crítico** | Datos personales expuestos a terceros externos | Fuga de MACs/IPs a Internet, acceso remoto al servidor | **1 hora** |
| **Alto** | Acceso no autorizado a datos pero sin exfiltración confirmada | SQL injection exitoso, acceso root al servidor | **4 horas** |
| **Medio** | Compromiso parcial sin acceso a datos personales | Credenciales admin comprometidas pero sin uso, DoS al portal | **24 horas** |
| **Bajo** | Evento sospechoso sin impacto confirmado | Intentos fallidos de login, anomalías en logs | **72 horas** |

---

## 3. Procedimiento de respuesta

### Fase 1: Detección y reporte (inmediato)

**Quién reporta**: Cualquier persona que detecte una anomalía.

**Cómo reportar**:
1. Contactar al administrador de red por correo: [correo del administrador]
2. Si no hay respuesta en 1 hora, escalar a: [correo de respaldo]
3. Incluir en el reporte:
   - Fecha y hora del evento detectado
   - Descripción de la anomalía
   - Archivos de log o evidencia disponibles
   - IP o identificador del sistema afectado

**Señales de alerta**:
- Logs con IPs desconocidas accediendo al panel admin
- Reglas iptables no autorizadas
- Archivo de base de datos modificado externamente
- Tráfico de red anómalo desde el servidor
- Alertas de integridad de archivos

### Fase 2: Contención (0-1 hora)

**Objetivo**: Detener la propagación del incidente.

| Acción | Comando / Procedimiento |
|--------|------------------------|
| Aislar el servidor de la red | Desconectar cable de red / desactivar interfaz |
| Bloquear IP sospechosa | `iptables -A INPUT -s <IP> -j DROP` |
| Revocar sesión de admin | Cambiar contraseña en `config.toml` + reiniciar |
| Detener el servicio | `systemctl stop catsplash` o `kill <PID>` |
| Preservar evidencia | `cp /var/log/syslog /evidencia/incidente_$(date +%s).log` |

### Fase 3: Evaluación (1-4 horas)

**Objetivo**: Determinar el alcance del daño.

| Pregunta | Cómo verificarlo |
|----------|-----------------|
| ¿Qué datos se expusieron? | Revisar `audit_log` → `ListAuditEvents()` |
| ¿Cuántos usuarios afectados? | Contar registros en `clients` al momento del incidente |
| ¿Hubo exfiltración? | Revisar tráfico de red, logs de firewall |
| ¿Qué vectores se explotaron? | Revisar logs de la aplicación y del sistema |
| ¿Hay persistencia del atacante? | Buscar usuarios, cron jobs, archivos modificados |

**Herramientas de investigación**:
- `audit_log` en SQLite: `SELECT * FROM audit_log WHERE timestamp > ?`
- Logs de la aplicación: `/var/log/catsplash.log`
- Logs del sistema: `/var/log/syslog`, `/var/log/auth.log`
- Reglas de firewall actuales: `iptables -L -n -v`
- Conexiones activas: `ss -tunap`

### Fase 4: Eradicación (4-24 horas)

**Objetivo**: Eliminar la causa raíz.

| Causa raíz | Acción de eradicação |
|------------|---------------------|
| Credenciales comprometidas | Regenerar hash bcrypt: `catsctl setpass <nueva>` |
| Vulnerabilidad en código | parchar y desplegar versión corregida |
| Configuración insegura | Revisar `config.toml`, permisos de archivos |
| Acceso físico no autorizado | Cambiar contraseñas, revisar permisos del OS |
| Software desactualizado | Actualizar dependencias: `go get -u ./...` |

### Fase 5: Notificación (24-72 horas)

#### 5.1 Notificación al titular afectado (Art. 50 LOPDP)

Si el incidente afecta datos personales de usuarios identificables:

| Campo | Contenido |
|-------|-----------|
| **A quién** | Todos los titulares cuyos datos pudieron verse comprometidos |
| **Medio** | Correo electrónico si está disponible, o aviso público en el portal |
| **Plazo** | Dentro de las 72 horas posteriores a la confirmación del incidente |
| **Contenido** | Naturaleza del incidente, datos afectados, medidas tomadas, recomendaciones |

**Plantilla de notificación**:

```
Asunto: Notificación de incidente de seguridad — Catsplash

Estimado/a usuario/a,

Le informamos que el [FECHA] se detectó un incidente de seguridad que pudo
haber afectado sus datos personales tratados por el sistema Catsplash.

Datos potencialmente comprometidos: [descripción específica]
Período del incidente: [inicio] - [fin]
Medidas adoptadas: [lista de acciones]

Recomendaciones:
- [recomendación 1]
- [recomendación 2]

Para más información, contacte a: [correo de contacto]

Lamentamos los inconvenientes.
```

#### 5.2 Notificación a la autoridad de control

Si el incidente afecta a un número significativo de titulares o involucra datos de alta sensibilidad, notificar a la **Superintendencia de Protección de Datos Personales** del Ecuador.

#### 5.3 Notificación interna

| Destinatario | Contenido |
|-------------|-----------|
| Administrador de red | Reporte completo del incidente |
| Desarrollador principal | Detalles técnicos para corrección |
| [Otros stakeholders] | Resumen ejecutivo |

### Fase 6: Lecciones aprendidas (1 semana después)

**Objetivo**: Prevenir la recurrencia.

| Actividad | Responsable | Plazo |
|-----------|-------------|-------|
| Documentar cronología completa del incidente | Administrador | 3 días |
| Identificar fallas en controles existentes | Desarrollador | 5 días |
| Implementar mejoras técnicas | Desarrollador | 2 semanas |
| Actualizar este plan de respuesta | Administrador | 2 semanas |
| Realizar simulacro (tabletop exercise) | Todos | 1 mes |

---

## 4. Roles y responsabilidades

| Rol | Responsabilidades |
|-----|-------------------|
| **Administrador de red** | Detección, contención, notificación, comunicación |
| **Desarrollador** | Evaluación técnica, erradicación, parches |
| **Responsable de cumplimiento** | Notificación a autoridad, documentación legal |

---

## 5. Contactos de emergencia

| Rol | Nombre | Contacto |
|-----|--------|----------|
| Administrador de red | [Nombre] | [correo / teléfono] |
| Desarrollador principal | [Nombre] | christopher.villamarin@protonmail.com |
| Autoridad de control | Superintendencia de Protección de Datos Personales | [sitio web oficial] |

---

## 6. Registro de incidentes

| # | Fecha | Nivel | Descripción | Estado | Lecciones |
|---|-------|-------|-------------|--------|-----------|
| — | — | — | (Sin incidentes registrados) | — | — |

---

## 7. Revisión y simulacros

| Actividad | Frecuencia |
|-----------|------------|
| Revisión de este plan | Cada 6 meses |
| Simulacro de respuesta | Cada 12 meses |
| Actualización de contactos | Cada 3 meses |
| Prueba de backups | Cada 1 mes |
