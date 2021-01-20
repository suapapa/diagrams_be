from diagrams import Diagram, Cluster
from diagrams.programming.framework import Vue
from diagrams.programming.language import Go
from diagrams.programming.language import Python

with Diagram("Diagrams Sandbox", show=False):
    with Cluster("gVisor"):
        diagrams = Python("diagrams")
        
    Vue("diagrams-front") >> Go("diagrams-backend") >> diagrams
