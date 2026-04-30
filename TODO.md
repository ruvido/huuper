# Battleplan 

Nella lista, il battleplan active deve avere una grafica diversa, sfondo giallo e scritte nere. Essere in cima a tutto ed essere completamente visibile. Poi tutta la lista dei battle plan completed ed infine gli archived. Ognuno di questi gruppi deve essere in ordine cronologico, chiaramente. Non c'è bisogno di separarli: semplicemente prima va active, poi completed, poi archived. 

Quando si crea un nuovo battle plan con il tasto +, se esiste un battle plan attivo -> bottom menu modal che dice: "Vuoi abbandonare il tuo attuale battleplan?" bottoni no yes (giallo) -> se yes archive il battleplan attivo e inizia il wizard per il nuovo battleplan 

Quando si finisce di creare il battleplan e si scrive sul database, bisogna tornare alla lista, non a view/id

Duration deve stare in alto a sinistra, nella stessa riga di visibility. Quindi in alto appare il progress bar poi sotto il progress bar, a sinistra duration, a destra visibility. 

Perché usi http://localhost:9090/admin/battleplan/view/?view=9302l3o7m5b4us5 e non view/id?? lo stesso per edit perche? sembra inutile e logorroico

---

Le routine dovrebbero essere salvate in una collection a parte. Nei campi ci deve essere anche relation to user. Un altro campo deve contenere l'array di quanti user stanno usando quella routine. Perché le routine possono essere condivisibili cioè leggibili da tutti. 
