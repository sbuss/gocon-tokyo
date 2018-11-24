# slide 1

Hi, I'm Steven Buss. I'm a software engineer at Google in San Francisco, and I
work on the Go runtime for App Engine Standard.

Thanks so much for having me, I love Japan and I'm really happy to be here!

I'm going to give you a brief history of Go on Google App Engine, including
some difficulties we faced. I'll then tell you all about our new second
generation runtime which lets you run regular, unmodified Go programs on the
cloud.

# slide 2

So, Act 1: The beginning.

# slide 3

App Engine was the first serverless platform. We started with Python in 2008
and then added Go in May 2011. That was version r57.2, which obviously predated
Go 1.0. It launched with a number of limitations, similar to the limitations in
the original Python runtime:

# slide 4

- No unsafe or sockets
- You didn't get a real filesystem: Because apps ran as a sandboxed process
  which didn't have permission to access the underlying filesystem, we patched
  in a virtual filesystem for a few whitelisted paths, but most operations
  failed with EPERM.
- We replaced syscall with a very limited surface
- And we didn't expose raw sockets, so we removed a lot of functionality from
  the net package.
- And, finally, we limited programs to a single thread which we optimistically
  hoped we would fix soon.
  Some background on that limitation: this was kind of inherited from the
  original design of the Python runtime. The execution and billing models all
  assumed that one request resulted in one thread being spawned. This was also
  partially to prevent inadvertently DOSing our servers from poorly written
  or malicious apps. Every request maps cleanly to a single thread.

# slide 5

App Engine's purpose is to run untrusted code on Google servers with no virtual
machine necessary. What could possibly go wrong?

Well, a lot, so in addition to limiting the standard library of Go, we ran apps
in a multi-tier sandbox to keep our customers safe.

# slide 6

The Python runtime pioneered this clever approach: in addition to hardware and
network isolation at the datacenter level, we also used Chrome's NaCl sandbox
and ptrace to provide two levels of security at the system level.

[3 minutes]

NaCl was developed by the Google Chrome team to securely run compiled
extensions on users' machines. It requires that developers link against NaCl,
which gives programs limited access to system resources. It was really quite
excellent at the time, but it's also kind of a headache to develop with. We
aren't using this anymore with the second-generation runtime, instead we're
using gVisor which I'll talk about later on.

ptrace is part of the linux kernel, and is a syscall that gives a program the
ability to intercept and modify the syscalls of another program. This is most
commonly used for debugging, where you have one program, the debugger,
controlling and modifying the execution of another program. You can also use it
to enforce that a program doesn't use any disallowed syscalls. As you might
expect, this makes programs execute quite a bit slower since every syscall
is intercepted and checked against a blacklist.

# slide 7

In order to run Go programs under NaCl and in our limited environment, we had
to patch the standard lib and build with CGO.

For example, I mentioned the virtual filesystem earlier. Well, it's not a
pollable filesystem and Go 1.9 combined the net/poll and filesystem packages
into a shared interface. This caused panics in our runtime and was one of the
many things I had to modify to run correctly.

We also disabled a lot of syscalls. We only allowed a small subset of syscalls
we deemed safe, and patched Go to use an internal syscall package specific to
App Engine. If you play around with the runtime you'll discover that things
like chmod, rmdir, mkdir, kill, and many other syscalls return EPERM.

The network isolation I'm sure caused you all plenty of headaches. We didn't
allow customer apps to open raw sockets, instead we intercepted socket calls
and redirected them. Network calls went from your app, over a unix domain
socket to the supervisor monitoring your app, then through a network proxy
which allowed access to internal APIs. This enabled cool things like automatic
authorization for API calls, but also meant you had to use the urlfetch package
rather than net/http.Get to access outside resources. Any access from your app
to the internet had to go through the urlfetch service. In addition, all
inbound traffic was required to be HTTP, and combined with the inability to
open a raw socket, prevented you from running a GRPC server on App Engine.

Unfortunately, we haven't lifted the HTTP requirement yet in the second
generation runtime, so inbound traffic must still be HTTP, but we are working
to fix that.

# slide 8

[6:30 minutes]

As I mentioned already, we use ptrace to monitor all of the syscalls your app
makes, so here's a little demo to show what using the ptrace syscall looks
like.

First, I want to give a BIG thank you to Liz Rice, this example is a slightly
modified version of her code from her talk at GopherCon 2017 titled "A Go
Programmer's Guide to Syscalls". I really encourage you to go watch her talk
and read her code. She goes into much more detail than I am, and I learned a
good deal from it.

[CHANGE TO ptrace-example]
[give demo. show the helloworld program, show how ptrace can intercept every
syscall, and show how a syscall can be filtered]

[ 16:30 minutes ]

# slide 9

Beyond the security features, we also wanted to provide a good user experience.
This meant a few key features:

- No need for package main Remember that we launched before Go 1.0, and before
  many best practices had emerged, so not requiring package main wasn't so
  crazy at the time. Instead, you would register your HTTP listeners in various
  init() funcs and we would synthesize a main package for you.

This is neat, but it also meant that go-app-builder diverged over time from `go
build`. This is the primary reason that we never got support for vendoring. I
tried several times to get vendoring working, but ultimately failed. There was
no hope of getting go-app-builder to support go modules, so it's quite lucky
that our second generation runtime was ready for Go 1.11.

- Also, we support apps that aren't on the GOPATH. In fact, go on App Engine
  predates both cmd/go and GOPATH. We *still* don't require you to develop
  inside GOPATH. Even with Go 1.11, you can use go modules which don't need to
  be on GOPATH either.

- We added the appengine package to the standard library which gave easy access
  to the App Engine APIs, which also meant we had to distribute a custom
  version of Go, which isn't great.
- You could call this a feature or not: we didn't support
  net/http.ListenAndServe because we didn't have a real network stack, instead
  we set up the webserver on your behalf. If you used App Engine prior to Go
  1.11, you probably used appengine.Main(), which set up the appropriate
  networking inside App Engine. If you are using Go 1.11, you can finally use
  ListenAndServe and everything works normally.

[ 19:40 ]

# slide 10

Go on App Engine had a great start back in 2011. But Go has changed a lot since
then, and we haven't always done a good job keeping up. Act 2: The journey

# slide 11

The original development team did such an amazing job adding features from 2011
to 2015. They added support for 
- all of the App Engine APIs,
- open sourced the appengine package as google.golang.org/appengine and
  encouraged users to migrate to it
- various improvements to the sandboxing and performance of the runtime
- and several version upgrades of Go

# slide 12

In March 2012 Go 1.0 launched on App Engine alongside the release of 1.0.
Sadly, that was the first of only two times we launched a new Go version
alongside the golden release.

And as many of you already know, we entirely skipped Go 1.3, 1.5, 1.7, and
1.10. Believe me when I say it wasn't for a lack of trying.

# slide 13

So why did we miss so many? There are many reasons:

- The biggest reason is that the NaCl environment is difficult, and maintaining
  the patches across Go versions while ensuring the security of the platform is
  a big undertaking. For example, I missed Go 1.7 because the patches from 1.6
  didnt apply cleanly. It required several weeks of work to get the runtime
  operational with 1.7, and then I had to track down random panics. By the
  time I got everything operational, Go 1.8 was only a month away so we opted
  to skip 1.7 and just get 1.8 out the door.
- Even when everything goes well, we still have to build on the golden release,
  which means we have to wait until it's ready before we start validating the
  runtime. That guarantees at least a month delay between a Go release and a
  launch on App Engine.
- We missed 1.10 because after Go 1.9 we decided to focus all new development
  on the new second-generation runtime, which meant I hadn't even done the
  work to get 1.10 running in the old environment. So when the second-gen
  runtime environment wasn't quite ready when Go 1.10 launched, we had no
  choice but to wait.

The good news is that with Go 1.11, adding new versions is as simple as
changing a config file, running the build, and then going through the internal
validation. It should dramatically reduce the delay between the golden Go
releases and App Engine Go releases.

[ 22:40 ]

# slide 14

In addition to all that, sometimes the Go 1 compatibility promise doesn't quite
live up to its goals. The Go 1.4 to 1.6 ugprade broke some customers, which
halted development until we added support for pinning to a Go version. That was
my first feature when I joined the team in 2016. It turns out that silently
upgrading your Go version isn't something you want all the time. Our next
generation runtimes currently pin you to a version by default.

# slide 15
[ 23:30 ]

So that's how we got here, what's ahead? Act 3: The Next Generation

[11 minutes]

# slide 16

The biggest change behind the scenes, and the one I'm most excited about, is
the move away from NaCl and ptrace to gVisor.

# slide 17

gVisor is a user-space kernel, written in Go, that implements a substantial
portion of the Linux system surface. Like ptrace, It intercepts application
system calls and acts as the guest kernel. Unlike our ptrace environment,
where we had to patch the golang syscall package to reimplement the syscalls
we needed, gVisor allows us to safely use any syscall that it implements in
a natural way.

In addition to a real kernel, you also get a real filesystem. So the filesystem
syscalls which I mentioned were disabled before (chmod, mkdir, etcetera) now
work correctly in a totally isolated environment. Note, though, that Your
filesystem only exists temporarily, so any writes you do will be discarded
after some time. This is still a serverless environment, after all, so If you
want to persist data you need to write it to something like Datastore or Google
Cloud Storage.

gVisor enables us to securely run untrusted code on our infrastructure without
building against any special libraries. So rather than using a custom version
of Go with a patched standard library and a custom go-app-builder, we can now
just run a stock `go build` on your code.

gVisor is super cool, and I encourage all of you to check it out at
github.com/google/gvisor

# slide 18

Ok, so gVisor is better than ptrace and NaCl, but why is it better than a
virtual machine?  It comes down to resource usage and an extensible
architecture. gVisor can operate on machines with limited cpu and ram; or taken
another way, it means that we can run way more apps on a given machine than we
could if we were running virtual machines. This leads to higher performance and
lower cost than you would otherwise get. I'll talk a bit more about the
extensible architecture in a couple slides.

Not everything is roses, though. gVisor has a few drawbacks compared to a
virtual machine: foremost is reduced compatibility. A virtual machine is
running a real linux kernel and supports anything linux does, whereas gVisor is
a guest kernel and doesn't implement everything. In practice, this probably
won't be a problem, but the gVisor team is constantly working to expand its
syscall coverage. It also has slightly higher syscall overhead than a VM, but
it's no worse than the old generation ptrace environment. Most relevant to you
as developers, some of the failures you might experience from an unimplemented
or buggy syscall might be hard to debug. It will keep improving, though.

# slide 19
[ 27:00 ]

The basic idea of gVisor is simple: run an untrusted binary on a guest kernel
which acts as a sandbox. Depending on the syscall, gVisor either reimplements
it or passes it through to the base kernel. This kind of functions like ptrace,
in that all syscalls can be inspected, but rather than killing a process or
returning EPERM when it uses a dangerous syscall, it instead gets a safe
reimplementation that only runs inside the security sandbox.

[ Demo ptrace helloworld running under gVisor. Show logs in /tmp/runsc ]


# slide 20

[ 31:00 ]
The extensible architecture is quite clever. devices and files can be
implemented as 'gofers'. These isolate external communication from the secure
kernel. This allows both security isolation so that, for example, mistakes in
net stack don’t leak into kernel space, and enables special file handles like
`/cloudsql` in App Engine which gives easy access to Google Cloud SQL. The
cloudsql gofer handles authentication and exposes a unix-socket interface to
your app which actually goes out over our internal network to your cloudsql
database.

Under gVisor, rather than patch the net and os packages, we can instead run
unmodified Go, with syscalls handled by the gVIsor sentry, and net and
file access handled by the Gofers. I really can't praise gVisor enough; you
should all go check it out, you can build really incredible things wth it.

Again, it's at github.com/google/gvisor

# slide 21

Ok, so gVisor is awesome -- how does that translate into a better experience on
App Engine?

First let's look at the old way of writing a service for Go on App Engine:

The old way is straightforward: import google.golang.org/appengine, call
appengine.Main(), and then things work. The downside of this is that, in order
to run your app locally, you had to use dev\_appserver.py or goapp.

# slide 22

The new way gets rid of all that. You set up your webserver exactly like you
would on any other platform. import net/http, call ListenAndServe on PORT, and
you're good to go.

[ Demo the old way and the new way.
1. show the old helloworld. highlight appengine.Main. Show running it without
   dev_appserver, and that it fails.
2. show the new helloworld. Show running locally. Show app.yaml. Show
   deployment.
3. Show old vendor failing. Explain problem with the builder?
4. Show new vendor succeeding.
5. Show new go.mod helloworld succeeding.]

# slide 23

[ 36:00 ]

Ok, so you get all this good stuff, what's the catch? Well... you should start
migrating away from the legacy App Engine APIs and toward the cloud APIs.

The legacy APIs are great, but they are a walled garden: you have to use App
Engine to use them. That means that you might feel locked-in to our platform.

The Google Cloud APIs provide platform-independent access to all of our
services, like Datastore and Cloud Machine Learning. There's no requirement
that you run your code on Google Cloud or App Engine, but I do think it's the
easiest and best product.

Migrating your code to use the Cloud APIs rather than the App Engine APIs will
take a little work, but it will future proof your app and make it possible
to seamlessly shift from App Engine to Kubernetes or virtual machines and
back as needed.

# slide 24

check out these URLs to learn how to migrate. Or, you know, just look on
stack overflow.

# slide 25

I'm really excited to be working on this runtime, and to be working on Google
Cloud. I love Go and the future of cloud at Google is built on it. All of these
products are written in go: from gVisor, the core virtualization technology for
App Engine second gen, to the runtime itself. Google Cloud Functions is
launching the Go runtime soon -- and check out the demo from GopherCon 2018 in
Denver. Kelsey Hightower gave an amazing demo of this excellent product. Not to
mention Kubernetes and Knative -- both written in Go.

There are a ton more things written in Go that power Google Cloud. I wish I
could talk about them all, but I'm glad I got to give you a little peek behind
the scenes. I'll be around today and I'll be happy to answer any questions you
have. Please feel free to pull me aside.

# slide 26

Oh, I almost forgot! It may have taken seven years but GOMAXPROCS is no longer
hardcoded to 1! Go write some concurrent programs!

# slide 27

Thank you so much!

[ 39:45 ]








