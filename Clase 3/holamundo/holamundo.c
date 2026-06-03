#include <linux/init.h> // Este archivo contiene las macros __init y __exit
/*
    Que son los macro?
    Los macros son una forma de definir constantes en C.
    En este caso, __init y __exit son macros que se utilizan para indicarle al kernel que funciones
    se deben llamar al cargar y descargar el modulo.
*/
#include <linux/module.h> // Este archivo contiene las funciones necesarias para la creacion de un modulo
#include <linux/kernel.h> // Este archivo contiene las funciones necesarias para la impresion de mensajes en el kernel

/*
    El modulo debe tener una licencia, una descripcion y un autor.
*/
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("A simple Hello, World Module");
MODULE_AUTHOR("Camilo");

static int __init hello_init(void) {
    printk(KERN_INFO "Hello, World!\n");
    return 0;
}

static void __exit hello_exit(void) {
    printk(KERN_INFO "Goodbye, World!\n");
}

/*
    Se debe indicarle al kernel que funciones se deben llamar al cargar y descargar el modulo.
*/
module_init(hello_init);
module_exit(hello_exit);