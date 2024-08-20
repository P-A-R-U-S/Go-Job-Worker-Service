## Questions for Kick-Off meeting:

### General:
- I have two day camping with family between 27-28 August. 
  Is it possible to extend challenge for 2 days ? I not sure, I would have any internet access or able to work from the campground. 

- My plan to start with Level 4 and if I would have extra time I would try to implement logic for: "Add resource isolation for using PID, mount, and networking namespaces.". 
  Can I do this, or I need to stick directly with Level 4 ?

- As far as assigment simulate day-as-usual communication we need to agreed when and how to provide status and what questions I can solve with team. 

- Challenge mentioned I need so submit - "roughly 3-5 Pull Requests". As I understand it would be
  1. Design Doc.
  2. Job Worker Library
  3. gRPC Service
  4. CLI
  5. Possible bug fixes.
  

- **Development environment:**
  I’m using MacOS as my primary developer box, but because assignment required to run Linux Kernel Base OS and support features like groups and etc, 
  I need to separate dev environment and I’ve two option - Linux based VM (KVM) running on QEMU/VirtualBox or host it somewhere (e.g. AWS or etc). 
  Are there any preference for the team ? 
  In case of hosting I need setup deployment process (probably with Terraform) and support it, but maybe team has already recommendation or preference what to do in this case. 
  For now plan use QEMU with Linux Kernel (Ubuntu).  

- Can I use external packages ? E.g. for groups ?

- Should we persist command output ? DataBase or File ? For simplicity, I would consider File is best option.

- CGroups: Controls groups are widely used running container and requirement mentioned that _**"The server should also not rely on any shell scripts, external binaries or use containers to execute jobs."**_
  Just like to confirm team are not expecting to run every job/command in separate docker container or etc, 
  but on separate goroutine. 

- For TLS setup: can I generate my own certificates ? Should I use publicly trusted certificate authority - e.g. DigiCert or etc. ?

- As I see for the purpose to complete on-time we are fine to pre-generate and store certificates in main repository.

- As I see testing should be done mostly for: _**"authentication/authorization layer and networking"**_. 
  Do you prefer unit or integration test ? For flexible design and good unit test coverage I would need to use mocking and design in this case would be heavily based on interfaces. 
  Are this acceptable or better to avoid this for simplicity ?

- CLI: Can I use external library (e.g. "github.com/urfave/cli" ) for handle CLI parameters ? 
  It would make CLI part much simpler.

- About streaming command output via gRPC:
  1. Can (or Should) I used streaming or can request output by chunks?
  2. Can I have little more details about - _**"Multiple concurrent clients should be supported."**_ ?

- CLI (authorization): CLI need to configure how connect to the Service. 
   Should I hardcode or store this authorization info? 