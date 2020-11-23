(module
    (import "env" "malloc" (func $gmalloc (param i32) (result i32)))
    (import "env" "free" (func $gfree (param i32)))
    (global $pointer (mut i32) (i32.const 0))
    (global $size i32 (i32.const 800))
    (memory (export "memory") 1)
    (data (offset i32.const 100) "yes it is")
    (func (export "thunderchain_main")
        global.get $size
        call $gmalloc
        global.set $pointer
        global.get $pointer
        call $gfree
    )
)