# KubeVirt en GKE desde Windows (PowerShell)
Kubevirt es una extensión de código abierto que permite crear, ejecutar y administrar máquinas virtuales (VMs) directamente dentro de un clúster de Kubernetes

Algunas cosas a tomar en cuenta son:
- El clúster GKE debe tener las máuqinas configuradas con tipo **N1 con virtualización anidada habilitada**
- Docker Desktop para Windows instalado
- Cuenta en Docker Hub

---

## Instalar KubeVirt (operator)

```powershell
# Obtener la versión estable
$VERSION = (Invoke-WebRequest -Uri "https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt").Content.Trim()

# Descargar el operator YAML
Invoke-WebRequest -Uri "https://github.com/kubevirt/kubevirt/releases/download/$VERSION/kubevirt-operator.yaml" -OutFile "kubevirt-operator.yaml"
```

> **Nota:** En GKE no se tiene acceso a los nodos `control-plane`, por lo que el operator fallará si busca `affinity`/`tolerations` de control-plane. Es necesario editar el archivo `kubevirt-operator.yaml` y eliminar los bloques `affinity` y `tolerations` que estan en los Deployments del operator.

Con los cambios realizados hacemos un create:

```bash
# Creacion
kubectl create -f kubevirt-operator.yaml

# Verificacion
kubectl describe pod -n kubevirt
```

En caso de que kubevirt no logre levantarse, quizá sea por el affinity y tolerations, para no modificar el archivo desde 0, se pueden usar los siguiente comandos:
```bash

# Eliminar affinities si quedaron:
kubectl patch deployment virt-operator -n kubevirt --type=json -p='[{"op": "remove", "path": "/spec/template/spec/affinity"}]'

# Eliminar tolerations si quedaron:
kubectl patch deployment virt-operator -n kubevirt --type=json -p='[{"op": "remove", "path": "/spec/template/spec/tolerations"}]'

# Verificar
kubectl get pods -n kubevirt -w
```

---

## Crear el Custom Resource de KubeVirt

Guardamos lo siguiente como `kubevirt-cr.yaml`:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  certificateRotateStrategy: {}
  configuration:
    developerConfiguration:
      featureGates: []
  customizeComponents: {}
  imagePullPolicy: IfNotPresent
  workloadUpdateStrategy: {}
  infra:
    nodePlacement: {}
  workloads:
    nodePlacement: {}
```

Hacemos el apply:

```powershell
kubectl apply -f kubevirt-cr.yml

# Verificar que todo esté corriendo
kubectl get all -n kubevirt
```

---

## Crear la imagen qcow2 con QEMU (desde WSL2)

QEMU no corre nativamente en Windows, así que será necesario **WSL2** (Ubuntu), dentro de wsl:

```bash
# Insalar wsl(ubuntu)
wsl --install

# Actualizar a wsl2
wsl --set-default-version 2

# Iniciar wsl
wsl

# Instalar QEMU
sudo apt update && sudo apt install -y qemu-system-x86 qemu-utils libvirt-daemon-system libvirt-clients bridge-utils

# Crear disco virtual (Acá se deine el espacio de "disco" duro de la vm en el último parámetro)
qemu-img create -f qcow2 alpine_disk.qcow2 5G

# Descargar ISO de Alpine
wget https://dl-cdn.alpinelinux.org/alpine/v3.23/releases/x86_64/alpine-standard-3.23.2-x86_64.iso

# Comprobar kvm
kvm-ok

# Instalar Alpine en el disco (arranca desde ISO) - Si kvm es soportado
sudo qemu-system-x86_64 -enable-kvm -m 2G -cpu host -smp 2 -cdrom alpine-standard-3.23.2-x86_64.iso -drive file=alpine_disk.qcow2,format=qcow2 -boot d

# Instalar Alpine en el disco (arranca desde ISO) - Si kvm NO es soportado
sudo qemu-system-x86_64 -m 2G -smp 2 -cdrom alpine-standard-3.23.2-x86_64.iso -drive file=alpine_disk.qcow2,format=qcow2 -boot d

```
Cuando inice en login escribimos root, luego dentro de alpine escribimos `setup-alpine` e iniciamos la instalación, la configuración estándar es:
- Keyboard: us / us
- Hostname: lo que queramos o Enter para default
- Network: eth0, DHCP, Enter
- Root password: ponemos cualquiera
- Timezone: America/Guatemala
- Proxy: Enter
- NTP Client: Enter
- APK Mirror: f
- Setup User: no
- SSH server: openssh
- Allow root ssh login?: yes
- Enter ssh key or URL for....: none
- Cuando pregunte Which disk(s) would you like to use? → escribimos sda
- Cuando pregunte How would you like to use it? → escribimos sys
- Erase above disk?: y

Una vez configurado apagamos la VM escribiendo `poweroff`
```bash
# Arrancar desde disco (sin ISO) para verificar y personalizar - si kvm es soportado
sudo qemu-system-x86_64 -enable-kvm -m 2G -cpu host -smp 2 \
  -drive file=alpine_disk.qcow2,format=qcow2

# Arrancar desde disco (sin ISO) para verificar y personalizar - si kvm NO es soportado
sudo qemu-system-x86_64 -m 2G -smp 2 -drive file=alpine_disk.qcow2,format=qcow2
```

> Si WSL2 no tiene soporte KVM, omitir `-enable-kvm` y `-cpu host` (será más lento pero funciona).

---

## Empaquetar la imagen en un contenedor y subirla a Docker Hub

Primero copiamos el `.qcow2` desde WSL a Windows (en PowerShell):

```powershell
# El filesystem de WSL está disponible en \\wsl$\Ubuntu\
Copy-Item "\\wsl$\Ubuntu\home\<tu_usuario>\alpine_disk.qcow2" -Destination "."
```
> Este último paso aplica si no se trabajó en el directorio del proyecto o un directorio de windows y en su lugar se trabajo en el sistema de archivos generado por wsl

Creamos el `Dockerfile`:

```powershell
@"
FROM scratch
ADD --chown=107:107 alpine_disk.qcow2 /disk/
"@ | Out-File -Encoding ascii Dockerfile
```

Build y push:

```powershell
docker build -t <USUARIO_DOCKER>/alpineimage:latest .
docker push <USUARIO_DOCKER>/alpineimage:latest
```

---

## Crear la VirtualMachine en KubeVirt

Guardamos lo siguiente como `vm.yaml` (Es necesario reempalazar `<docker_hub_user>`):

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: alpinevm
spec:
  runStrategy: Halted
  template:
    metadata:
      labels:
        kubevirt.io/size: medium
        kubevirt.io/domain: alpineso1
    spec:
      domain:
        devices:
          disks:
            - name: containerdisk
              disk:
                bus: virtio
          interfaces:
          - name: default
            masquerade: {}
        cpu:
          cores: 1
        resources:
          requests:
            memory: 1Gi
      networks:
      - name: default
        pod: {}
      volumes:
        - name: containerdisk
          containerDisk:
            image: <docker_hub_user>/alpineimage
```

Hacemos el apply:

```powershell
kubectl apply -f vm.yml
kubectl get vms
```

---

## Instalar `virtctl` en Windows

```powershell
# Obtener la versión que coincide con el cluster
$VERSION = kubectl get kubevirt.kubevirt.io/kubevirt -n kubevirt -o=jsonpath="{.status.observedKubeVirtVersion}"

# Descargar virtctl para Windows
Invoke-WebRequest -Uri "https://github.com/kubevirt/kubevirt/releases/download/$VERSION/virtctl-$VERSION-windows-amd64.exe" -OutFile "virtctl.exe"
```

Con el .exe descargado hay 2 opciones:
- Mover el .exe al PATH para que sea mas comodo usar el CLI
- Dejarlo donde el directorio del proyecto y anteponer `.\virtctl.exe` a cualquier comando relacionado a virtctl

En caso de querer agregarlo a PATH, lo recomendable es colcoarlo en el mismo lugar que kubectl, para esto ejecutamos:

```powershell
# Ver dónde está kubectl para usar la misma carpeta
Get-Command kubectl | Select-Object -ExpandProperty Source
```

Esto útlimo nos dará un directorio y pegamos el .exe de virtctl en ese mismo lugar.

---

## Arrancar la VM y exponer servicios

```powershell
# Iniciar la VM (Es alpinevm porque así está definido en vm.yaml)
virtctl start alpinevm

# Verificar estado
kubectl get vms
kubectl get vmis   # VirtualMachineInstance (cuando está corriendo)

# Exponer un puerto (ej. SSH en 22)
virtctl expose vm alpinevm --port=22 --name=alpinevm-ssh --type=ClusterIP

# Acceder por consola serial
virtctl console alpinevm
```

# Instalación de Containerd + grafana
## Configuración de Containerd
Dentro de la VM de Alpine en KubeVirt, primero instalamos los paquetes necesarios:

```bash
# Actualizar repositorios y habilitar community repo
echo "https://dl-cdn.alpinelinux.org/alpine/v3.23/community" >> /etc/apk/repositories
apk update

# Instalamos containerd
apk add containerd
apk add containerd-ctr
```

Para iniciar containerd:

```bash
rc-service containerd start
rc-update add containerd default
```

## Descarga y Ejecución de Grafana con containerd

Primero hacemos pull y run de Grafana:
```bash
ctr image pull docker.io/grafana/grafana:latest
ctr run -d --net-host docker.io/grafana/grafana:latest grafana
```

Para poder acceder a Grafana desde fuera de la VM, en PowerShell exponemos el puerto con:
```bash
virtctl expose vm alpinevm --port=3000 --name=alpinevm-grafana --type=NodePort
```

Luego obtenemos el puerto asignado:
```bash
kubectl get svc alpinevm-grafana

#http://<EXTERNAL_IP_NODO>:<NODEPORT>
```

# Limpiar de Kubevirt (Completo)
Para eliminar todo lo relacionado con kubevirt

```bash
# Eliminar la VM
kubectl delete vm alpinevm

# Eliminar el Custom Resource de KubeVirt
kubectl delete -f kubevirt-cr.yml

# Eliminar el operator
kubectl delete -f kubevirt-operator.yml

# Verificar que el namespace kubevirt se eliminó
kubectl get ns kubevirt
```

Si el namespace queda en Terminating por mucho tiempo:
```bash
# Eliminar la VM
kubectl delete ns kubevirt --force --grace-period=0
```