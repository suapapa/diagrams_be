from diagrams import Cluster, Diagram
from diagrams.k8s.compute import Pod, Deploy
from diagrams.k8s.network import Service
from diagrams.k8s.infra import Master, Node
from diagrams.programming.language import Go
from diagrams.programming.framework import Vue
from diagrams.generic.database import SQL
from diagrams.generic.device import Tablet

print("hello world!\n")
with Diagram("", show=False):
    with Cluster("cluster"):
        Master("master") - Node("worker1") - Node("worker2")

    with Cluster("worker node"):
        with Cluster("deploy"):
            pod = [ Pod("hello"), Pod("hello"), Pod("hello")]
        nodeport = Service("NodePort")
        nodeport >> pod
    Tablet("user") >> nodeport

    with Cluster("container in Pod, hello"):
        Vue("front end") - Go("back end") - SQL("db")
