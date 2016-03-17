RELISH INSTALLATION AND BUILD INSTRUCTIONS - SOURCE DISTRIBUTION
================================================================



Steps for Linux and Apple Mac    
-----------------------------

1. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs. 
   (On Mac, you can type 'terminal' into the Spotlight search to launch a terminal window.)    

2. Install gcc (c compiler) so that go can build the go-sqlite library.

   On Mac, you can install the gcc compiler by downloading and installing the 
   <b>Command Line Developer Tools</b> subset of Xcode, which is 
   currently described and available at developer.apple.com/technologies/tools/features. 
   Note you need an AppleID to download XCode tools.

   On ubuntu linux, 
   sudo apt-get install build-essential 
   will install gcc if "which gcc" reveals it does not exist on your platform.   

3. Install git version control system on your machine if not already installed.
   
   For linux:
   
   aptitude show git
   ...and if State: not installed...   
   sudo apt-get install git

   or equivalent package installation procedure on your linux distribution.

4. Install Go 1.3 or higher (from http://golang.org). 
   A binary distribution is available for some linuxes including Ubuntu. 
   A package installer (.pkg file) is available for MACOSX. 

5. Choose where you will create your mercurial repository for relish

   A relish source code distribution should be located at either
   ~/devel/relish or /opt/devel/relish. Some relish tools in future
   may depend on these locations.
   
   So change directory (cd) to one of the following locations: 
   
   cd       # change to your home directory e.g. /home/eric 
   mkdir devel
   cd devel
   
   or

   cd /opt
   sudo mkdir devel
   sudo chmod go+w devel   
   cd devel   

6.  git clone https://github.com/eric-hawthorne/relish.git

   Note that the repository directory that you are developing in and running relish tools from must be called relish.
   If you want to have multiple clones, you can clone them into e.g. relish_mybranch then
   use a symbolic link to decide which one is active. e.g. cd /opt/devel ; ln -s relish_mybranch relish
   

   We will use /opt/devel/relish as an example relish home directory in further instruction steps


7. Tell Go where to be looking for Go source code and where to be installing packages and executables:
   E.g. edit ~/.bashrc (if in linux) (~/.bash_profile on MACOSX) and add the following line:

   export GOPATH=/opt/devel/relish

8. Open a new terminal window to have a shell that recognizes the new PATH, or else 
    source ~/.bashrc (~/.bash_profile) in your current terminal window.


9. Install the go-sqlite package into the Go environment.

   go get github.com/mxk/go-sqlite/sqlite3   

10. build the relish compiler-interpreter 

    go install relish/relish
      
11. Add relish commands to your PATH, and make them executable, using the install.sh script.

   cd /opt/devel/relish
   sudo ./install.sh

   Instead of running ./install.sh, you can instead choose to append /opt/devel/relish/bin to 
   your PATH environment variable in your ~/.bashrc file. (~/.bash_profile on MACOSX).   
   Oh, then also do chmod go+x /opt/relish/bin/* 

12. Issue a command to compile and run a relish program. 

    relish shared.relish.pl2012/hello_application

    Note that this will first download the shared.relish.pl/hello_application relish
    software artifact from http://shared.relish.pl, will verify the authenticity and
    uncorruptedness of the artifact, then will install it locally, compile its
    relish source code, load .rlc intermediate-code into the runtime environment,
    then execute the main function in the main package of the software artifact. 
   
    Executing this program should print the version of the relish interpreter,
    then print an international version of "Hello World".

    You can find the auto-installed code of the downloaded artifact under the directory
    /opt/devel/relish/rt/shared/relish/replicas/shared.relish.pl2012/hello_application/

13. To get and build an updated version of relish source, cd to your /opt/devel/relish directory 
    or a subdirectory, then: 
    
    git pull
    go install relish/relish
    
    
