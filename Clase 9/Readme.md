# Kubernetes (K8s)

> **Nota:**
> `Esta guía fue realizada utilizando la versión 1.18`

Kubernetes es un sistema de orquestación de contenedores que automatiza el despliegue, la gestión y la escalabilidad de aplicaciones en contenedores. A continuación se presenta una descripción general de Kubernetes, junto con algunos comandos y ejemplos útiles.

## Configuración de GKE
Para usar k8s en gcp seguimos los siguientes pasos:
- Abrimos la consola de gcp
- Abrimos el menú de opciones del lateral izquierdo y buscamos *Kubernetes Engine*
- Seleccionamos la opción de Clusteres
- Elegimos la opción "Estándar" dando click en *Configurar*
- Definimos un nombre para el clúster
- En *Tipo de Ubicación* seleccionamos *zonal* para reducir costos.
- Seleccionamos el pool de *Grupos de Nodos*:
    - Reducimos el tamaño o canitdad de nodos para reducir más los costos.
    - Seleccionamos la opción *Nodos* y elegimos la configuración **N1**
    - Reducimos el tamaño de arranque de los nodos (por temas de costos)

Para usar el clúster tenemos dos opciones:
- Usar la consola de google
- Conectarnos a través de la línea de comandos

Para usar la línea de comandos debemos instalar gcloud:
https://docs.cloud.google.com/sdk/docs/install-sdk?hl=es


## Conceptos Principales

### Cluster

Un cluster es un conjunto de máquinas (nodos) que ejecutan aplicaciones en contenedores, coordinados por un nodo maestro (Control Plane).

### Node

Cada node es una máquina (física o virtual) que forma parte del clúster y ejecuta los Pods.

### Pod

El Pod es la unidad mínima de trabajo en Kubernetes. Puede contener uno o más contenedores y comparte recursos como almacenamiento y red.

De manera técnica no decimos que hacemos deployments de contenedores directamente en k8s, hacemos deployments de pods y usamos cosas sobre el pod para controlarlo.

### Controlador

Sirve para crear o actualizar pods y otros objetos de k8s, casi siempre será necesario crear un controlador, en k8s podemos encontrar los siguientes tipos de controladores:

- **Deployments, ReplicaSet:** Son los que utilizarán casi siempre, controlan a los pods en niveles muy bajos.

### Service o Servicio

Es el final o punto final dado hacia un pod, al usar un deployment para un conjunto de pods, es en ese momento que se configura un servicio.

### Namespace

Un Namespace es un "filtro" para la línea de comandos, es para ver solo lo que nosotros queramos.

### ConfigMap y Secret

- ConfigMap: Almacena configuraciones no sensibles. 
- Secret: Almacena datos sensibles (como contraseñas) de forma segura.

---

## Kubectl



Es la herramienta de línea de comandos de k8s, que también se conoce como kube control.

Se utiliza para interactuar y administrar el clúster.

### Algunos comandos:

- `kubectl version` : Para ver la versión instalada de Kubectl
- `kubectl run` : Crear un pod
- `kubectl cluster-info` : Para ver información del cluster
- `kubectl get nodes` : Para ver qué nodos están disponibles en el clúster

Ejemplo simple:
```bash
kubectl run nginx --image nginx        # Crea un Pod con nombre nginx
kubectl create deployment nginx --image nginx  # Creamos un deployment con nombre de nginx
```

---


## PODS

Nuestro objetivo es desplegar nuestra aplicación en forma de contenedores, sin embargo, k8s no implementa directamente sobre los nodos, lo hace en pods (la unidad mínima).

Lo usual es que por cada pod exista un contenedor. Si decrece el número de usuarios, solo se elimina un pod, no se suelen añadir contenedores adicionales a un Pod existente para el escalamiento.

Sin embargo, k8s no limita tener un contenedor por pod, podemos tener una aplicación entera en un pod (no es lo normal) y escalar todo con la creación de pods.

### Creación de pods y relación con deployments

Empecemos creando un pod:

```bash
kubectl run nginx --image nginx
```

Para eliminar el pod:

```bash
kubectl delete pod <nombre-del-pod>
```

Esto solo crea el pod, no el deployment, si no creamos un Deployment en Kubernetes, los Pods se crean individualmente (o como un ReplicaSet no administrado), pero se pierden las funcionalidades clave de autocuración, escalado automático y actualizaciones declarativas.

El siguiente comando realiza un deployment de un contenedor de docker creando un pod de k8s:

```bash
kubectl create deployment nginx --image nginx  # Creamos un deployment con nombre de nginx
```

Para entender mejor, primero se crea un pod automáticamente y hace un deployment de una instancia de la imagen del docker de nginx. Tener en cuenta que para obtener la imagen, se necesita especificar el nombre usando `--image`.

Para ver el pod creado:

```bash
kubectl get pods          # Información general del pod
kubectl get pods -o wide  # Agregar ip y el nodo donde está disponible en la salida
```

Para ver el deployment:

```bash
kubectl get deployments
```

Para eliminar el pod y deployment:

```bash
kubectl delete deployment <nombre-del-deployment>
```

Cuando creamos un pod a través de un deployment, la eliminación del deployment automáticamente elimina el pod.

## PODs y YAML

### Logs

Existe el comando `kubectl logs` que nos mostrará todos los recursos relacionados a un pod (también se puede especificar otros objetos).

Ejemplo de Pingpong:

```bash
kubectl run nginx --image pingpong ping 1.1.1.1
kubectl logs pod/pingpong
```

### Pod con Yaml

Kubernetes usa yaml como "entradas" para crear sus diferentes objetos (pods, replicasets, deployments, etc). Para esto se definen 4 campos obligatorios:

- `apiVersion` : La versión de api de kubernetes a usar para crear objetos, esta varía según el objeto, para pods es v1.
- `kind` : Indica el tipo de objeto a crear
- `metadata` : Se indican cosas como names y labels
- `spec` : Se indican especificaciones específicas a k8s, como los contenedores para el pod.

Cuando tengamos un archivo Yaml para algún objeto usaremos el siguiente comando para crearlo:

```bash
kubectl create -f <nombre-del-yml>  # Se considera que el comando se ejecuta en el directorio del .yml
```

## Replication Controller

Recordemos que los *controladores* son el cerebro de k8s, ya que se encargan de monitorear los objetos para que actúen como se espera.

El replication controller se encarga y asegura que se ejecuten el número específico de pods que se le hayan indicado en todo momento, también sirve para balancear y escalar la aplicación.

Este controlar nos ayuda a ejecutar varias instancias de un mismo pod en un mismo clúster de k8s. En caso de que solo tengamos 1 pod, este controlador crea un nuevo pod en caso de que el anterior falle.

La estructura de un yml de este controlador cuenta con los mismos niveles, con la única diferencia que `spec` tiene un nuevo campo:

- `template` : Acá se indica el pod al cuál se le aplicará el controlador

Además también se agrega `replicas`, la cual sirve para indicar cuantas copias del pod se necesitan en todo momento para nuestra aplicación.

Ubicamos nuestra carpeta con el controlador y para "ejecutar" el archivo usamos:

```bash
kubectl apply -f ej2-rc.yml
```

En caso de que queramos ver nuestros controladores de réplicas, usamos:

```bash
kubectl get replicationcontroller
```

Para ver los pods creados por el replication controller:

```bash
kubectl get pods
```

Si borramos un pod, el replication controller creará uno nuevo, esto se puede probar con:

```bash
kubectl delete pod <nombre-del-pod>
```

Si volvemos a usar un get de los pods, entonces veremos que aún existe la cantidad de pods que indicamos, pero ya no estará el que eliminamos, será uno nuevo.

Para borrar un replication controller:

```bash
kubectl delete replicationcontroller <nombre-del-controller>
```

### Replication controller y replica set

- **Replication Controller (RC):** Utiliza únicamente selectores basados en igualdad (equality-based selectors). Esto significa que solo puede seleccionar Pods que tengan una etiqueta con un valor exacto y específico (ej: app=nginx).
- **Replica Set (RS):** Introduce selectores basados en conjuntos (set-based selectors), que son mucho más potentes y expresivos. Permite filtrar etiquetas utilizando operadores como In, NotIn, Exists o DoesNotExist (ej: app in (nginx, apache)).

---

## Deployments

Es un objeto de kubernetes en 1 nivel superior a los ReplicaSet. La función principal de un deployment es actualizar y hacer un upgrade a instancias subyacentes con actualizaciones continuas (rolling updates), también permite deshacer, pausar y reanudar cambios según la necesidad. Al crear automáticamente se agrega un ReplicaSet que a la vez crean los PODs de nuestra aplicación, algunas características de los deployments son:

### Rolling update

Con k8s vimos que podemos tener varias instancias de una aplicación, si nosotros tratamos de actualizar nuestra aplicación, aplicar los cambios a todas las instancias de golpe podría generar problemas a los usuarios. Rolling Update permite aplicar los cambios a una instancia tras otra.

### Rollback

Es simplemente deshacer cambios recientes y volver atrás.

### ¿Cómo funciona un Deployment?

¿Cómo funciona?

- **Etiquetado de Pods:** Cuando se define un Deployment, también se describe un template (plantilla) para los Pods que creará. En esa plantilla, se usan labels para asignar etiquetas (ej: app: mi-aplicacion, version: v1) a los Pods que se generen.
- **Selección por matchLabels:** La sección matchLabels en el Deployment especifica las etiquetas que debe buscar.
- **Control y Mantenimiento:** El controlador del Deployment (y su ReplicaSet) monitorea constantemente el clúster. Si encuentra Pods que coincidan con esas etiquetas matchLabels, los gestiona (los mantiene vivos, los actualiza o los elimina). Si un Pod con esas etiquetas desaparece, el Deployment creará uno nuevo para mantener el replicas deseado.

### Comandos para crear un deployment con un yml

Con nuestro yml de deploy creado, usamos el siguiente comando para hacer deploy:

```bash
kubectl apply -f <nombre-del-deploy>
```

Para verificar todo lo que el deploy crea usamos:

```bash
kubectl get all
```

Es posible cambiar la cantidad de replicas del replica set definido en el deploy con:

```bash
kubectl deploy/<nombre-del-deploy> --replicas <num-replicas-deseadas>
```

Para eliminar el deploy usamos:

```bash
kubectl delete deployment <nombre-del-deploy>
```

### Actualización de Deployments

Hay 2 formas de actualizar la versión de la aplicación, contenedores, labels, etc.

#### Recreate

Esta estrategia consiste en dar de baja todas las instancias de la aplicación por luego crear las nuevas, esto si bien funciona generará inactividad de la aplicación.

#### Rolling Update

En esta estrategia se da de baja y se crean las instancias de manera consecutiva, una instancia tras otra.

#### Rollback

Es el proceso de revertir una actualización si algo sale mal durante el proceso. Para esto, el deployment eliminará los pods nuevos uno por uno y traerá los pods de la versión anterior.

Para poner a prueba estos conceptos tendremos que modificar el archivo `ej3-deploy`, específicamente indicaremos la versión de nginx en imagen, en este caso usaremos la `1.17.3`.

Con la imagen definida usamos:

```bash
kubectl apply -f <nombre-del-deploy>
```

Recibimos el siguiente comando para revisar el deploy:

```bash
kubectl describe deployment <nombre-del-deploy>
```

Ahora usemos el siguiente comando para ver el estado del rollout:

```bash
kubectl rollout status deploy/<nombre-del-deploy>
```

Ahora revisar el historial del deploy:

```bash
kubectl rollout history deploy/<nombre-del-deploy>
```

No se verá un registro como tal, para que en un cambio se mire el registro o record se usa:

```bash
kubectl apply -f <nombre-del-deploy> --record
```

Para actualizar el deployment sin usar el yml, se usa:

```bash
kubectl set image deployment/<nombre-del-deploy> nginx-container=nginx:1.17.10
```

Revisemos el deployment con:

```bash
kubectl describe deployment
```

Si tuviéramos un error en nuestra nueva versión y necesitamos hacer un rollback, usamos:

```bash
kubectl rollout undo deployment/<nombre-del-deploy>
```

Para revisar las versiones:
```bash
kubectl rollout history deploy/<nombre-del-deploy>
```

## Servicios

Los servicios permiten comunicar diferentes componentes dentro y fuera de la aplicación.

Para la conexión a pods, se requiere de un servicio, en pocas palabras, un service o servicio es una dirección estable para un pod o grupo de pods.

En la práctica buscamos exponer servicios para crear estos recursos que apuntan a pods de backends. Esto normalmente se hace con `kubectl expose` o con archivos yaml.

Hay distintos tipos de servicios:

- **ClusterIP:** Funciona en cualquier configuración, expone el servicio en una *IP interna* del clúster. Es decir, solo se encuentra disponible dentro del clúster (entre nodos y pods).
- **NodePort:** Funciona en cualquier configuración, está diseñado para ser accesible desde fuera de nuestro clúster, disponible para todos los nodos y cualquiera que se quiera conectar a él. (Se tendrán puertos altos)
- **LoadBalancer:** Es usar un servicio externo de un tercero, puede ser un proxy o firewall (AWS, Azure, GCP). Acá se expone el servicio de manera externa con un proveedor cloud.
- **ExternalName:** Proporciona un alias interno para un nombre de DNS externo. Los clientes solicitan el DNS y las solicitudes se redireccionan a un nombre externo.


---

## ConfigMaps

### Archivos de ejemplo

```
configmaps/
├── configmap1.yaml          # pares clave-valor simples
├── configmap2.yaml            # archivos completos embebidos
└── configmap3.yaml  # deployment que los consume
```

### 1. ConfigMap con pares clave-valor (`configmap1.yml`)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  APP_ENV: "production"
  APP_PORT: "8080"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-svc.default.svc.cluster.local"
  DB_PORT: "5432"
  DB_NAME: "mi_base_de_datos"
```

Aplicar y verificar:

```bash
kubectl apply -f configmaps/configmap1.yaml

kubectl get configmap app-config
kubectl describe configmap app-config
```


---

### 2. ConfigMap con archivos embebidos (`configmap2.yml`)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-files
  namespace: default
data:
  app.properties: |
    server.port=8080
    log.level=info

  nginx.conf: |
    server {
      listen 80;
      location / {
        proxy_pass http://app-svc:8080;
      }
    }
```

Cada clave del `data` se convierte en un archivo independiente cuando se monta como volumen.

---

### 3. Deployment consumiendo ConfigMaps (`configmap3.yml`)

Hay tres formas de exponer un ConfigMap a un pod:

#### Opción A — Variable de entorno individual

```yaml
env:
- name: APP_ENV
  valueFrom:
    configMapKeyRef:
      name: app-config
      key: APP_ENV
```

#### Opción B — Todas las claves como variables de entorno

```yaml
envFrom:
- configMapRef:
    name: app-config
```

#### Opción C — Montar como archivos en un volumen

```yaml
volumeMounts:
- name: config-vol
  mountPath: /etc/nginx/conf.d
  readOnly: true

volumes:
- name: config-vol
  configMap:
    name: app-config-files
    items:
    - key: nginx.conf
      path: default.conf      # nombre del archivo resultante
```

Verificar el archivo dentro del pod:

```bash
POD=$(kubectl get pod -l app=mi-app -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD -- cat /etc/nginx/conf.d/default.conf
```

---

## Secrets

### Archivos de ejemplo

```
secrets/
├── secret1.yaml              # valores en base64 (tipo más común)
├── secret2.yaml          # valores en texto plano (k8s codifica solo)
├── secret3.yaml     # credenciales de registry privado
└── secret4.yaml     # deployment que los consume
```

### 1. Secret tipo Opaque con base64 (`secret1.yml`)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: default
type: Opaque
data:
  DB_USER: bWlfdXN1YXJpbw==         # "mi_usuario"
  DB_PASSWORD: c3VwZXJfc2VjcmV0XzEyMw==  # "super_secret_123"
  API_KEY: YWJjZGVmZ2hpamtsbW5vcA==
```

Generar el base64 de un valor:

```bash
echo -n "mi_usuario" | base64
# bWlfdXN1YXJpbw==

# Decodificar para verificar
echo "bWlfdXN1YXJpbw==" | base64 -d
# mi_usuario
```

> **Importante:** usar `echo -n` para evitar que el salto de línea quede codificado.

---

### 2. Secret con `stringData` (texto plano) (`secret2.yml`)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secret-plain
  namespace: default
type: Opaque
stringData:
  DB_USER: "mi_usuario"
  DB_PASSWORD: "super_secret_123"

  connection.env: |
    DB_HOST=postgres-svc.default.svc.cluster.local
    DB_PORT=5432
    DB_USER=mi_usuario
    DB_PASSWORD=super_secret_123
```

`stringData` acepta texto plano; Kubernetes lo codifica en base64 internamente. Es útil para legibilidad en desarrollo. **No se debe commitear en repositorios.**

---

### 3. Secret para registry privado (`secret3.yml`)

La forma más práctica es generarlo con `kubectl`:

```bash
kubectl create secret docker-registry registry-secret \
  --docker-server=registry.example.com \
  --docker-username=mi_usuario \
  --docker-password=mi_password \
  --dry-run=client -o yaml > secret3.yml

kubectl apply -f secrets/secret3.yaml
```

---

### 4. Deployment consumiendo Secrets (`secret4.yml`)

Las tres opciones son análogas a los ConfigMaps, usando `secretKeyRef` en lugar de `configMapKeyRef`:

#### Opción A — Variable individual

```yaml
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: app-secret
      key: DB_PASSWORD
```

#### Opción B — Todas las claves como variables

```yaml
envFrom:
- secretRef:
    name: app-secret
```

#### Opción C — Montar como archivo (recomendado para producción)

```yaml
volumeMounts:
- name: secret-vol
  mountPath: /etc/app/secrets
  readOnly: true

volumes:
- name: secret-vol
  secret:
    secretName: app-secret-plain
    items:
    - key: connection.env
      path: connection.env
    defaultMode: 0400             # solo lectura para el propietario
```

Montar como archivo es preferible a variables de entorno porque los archivos pueden actualizarse sin reiniciar el pod, y no aparecen en logs de procesos.

---

## Aplicar todos los ejemplos en orden

```bash
# 1. ConfigMaps
kubectl apply -f configmaps/configmap1.yml
kubectl apply -f configmaps/configmap2.yml
kubectl apply -f configmaps/configmap3.yaml

# 2. Secrets
kubectl apply -f secrets/secret1.yaml
kubectl apply -f secrets/secret2.yaml
kubectl apply -f secrets/secret3.yaml
kubectl apply -f secrets/secret4.yaml

# 3. Verificar
kubectl get configmaps
kubectl get secrets
kubectl get pods
```

---

## Comandos útiles de diagnóstico

```bash
# Ver contenido de un ConfigMap
kubectl get configmap app-config -o yaml

# Ver un Secret (los valores aparecen en base64)
kubectl get secret app-secret -o yaml

# Decodificar un valor específico de un Secret
kubectl get secret app-secret \
  -o jsonpath='{.data.DB_PASSWORD}' | base64 -d

# Verificar que las variables llegaron al pod
POD=$(kubectl get pod -l app=mi-app -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD -- env | grep -E "APP_|DB_"

# Verificar un archivo montado desde Secret
kubectl exec $POD -- cat /etc/app/secrets/connection.env
```

---

## Actualizar un ConfigMap o Secret sin borrar el recurso

```bash
# Editar directamente
kubectl edit configmap app-config
kubectl edit secret app-secret

# O re-aplicar el YAML modificado
kubectl apply -f configmaps/configmap1.yaml
```

> Los pods que consumen el recurso **como volumen** reciben el cambio automáticamente en ~1-2 minutos. Los que lo consumen **como variables de entorno** requieren reinicio del pod para reflejar cambios.
