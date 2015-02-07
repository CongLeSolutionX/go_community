subroutine print_hi() bind(C)
  implicit none
  write(*,*) "Hello from Fortran."
end subroutine print_hi
