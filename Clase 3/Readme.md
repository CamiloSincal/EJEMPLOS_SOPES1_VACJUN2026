# Creación de módulos en Linux
## Instalación/configuración de C y python en Linux

Para instalar C en Linux, se debe instalar el compilador de C, gcc. Para instalar gcc en Ubuntu o Fedora, se debe ejecutar el siguiente comando:

```bash
sudo apt-get install gcc
# ver versión
gcc --version
```

Además, dado que se usará **Makefile** para la compilación, si no tienen instalado el paquete make pueden instalarlo con el siguiente comando:

```bash
sudo apt-get install make
```

También será necesario instalar los essentials de desarrollo en Ubuntu, los essentials de desarrollo incluyen herramientas y bibliotecas necesarias para compilar programas en C, para ello se debe ejecutar el siguiente comando:

```bash
sudo apt-get install build-essential
```
## Instalación de Python

Python usualmente viene instalado en la mayoría de las distribuciones de Linux, sin embargo, se puede verificar si está instalado ejecutando el siguiente comando:

```bash
python --version
```

En caso de que no esté instalado, se puede instalar Python en Ubuntu o Fedora ejecutando el siguiente comando:

```bash
sudo apt-get/dnf install python3
```

## Pasos para compilar e instalar un módulo
```bash
make # se compila y genera los archivos de instalación
sudo insmod <file>.ko # Instala el módulo de kernel

sudo dmesg | tail - n 20 # para ver los últimos 20 logs del kernel

sudo rmmod <name> # Desinstalar el módulo

```

# Módulos para imprimir métricas del SO en un archivo /proc
Antes de crear y compilar este tipo de módulos es necesario preparar el entorno, para esto debemos:

- Asegurnarnos de tener instalado un compilador de kernel, como gcc.
- Tener acceso a los encabezados del kernel (kernel headers). En Ubuntu y sistemas basados en el mismo, pueden instalarlos con:

```bash
sudo apt install linux-headers-$(uname -r)
```

Si durante la compilación con `make` se genera un error, es posible que se tengan que generar los archivos de configuración, para eso usamos los siguientes comandos:

```bash
# Se navega al directorio de los headers
cd /usr/src/linux-headers-$(uname -r)

# Se copia la configuración actual del kernel
sudo cp /boot/config-$(uname -r) .config

# Se generan los archivos faltantes
sudo make oldconfig
sudo make prepare
sudo make modules_prepare
```

En último caso es posible que se necesite una reinstalación:

```bash
# Se remueven headers actuales
sudo apt remove linux-headers-6.14.0-36-generic

# Se limpia la configuración
sudo rm -rf /usr/src/linux-headers-6.14.0-36-generic

# Se actualiza e instala de nuevo
sudo apt update
sudo apt install --reinstall linux-headers-6.14.0-36-generic
```