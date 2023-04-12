!function(){"use strict";var e,t,n,o=window.location,c=window.document,u=c.getElementById("vince"),s=u.getAttribute("data-api")||(e=u.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function p(e){console.warn("Ignoring Event: "+e)}function r(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return p("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return p("localStorage flag")}catch(e){}var n={};n.n=e,n.u=t&&t.u?t.u:o.href,n.d=u.getAttribute("data-domain"),n.r=c.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props);var r=u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),a=n.p||{};r.forEach(function(e){var t=e.replace("event-",""),n=u.getAttribute(e);a[t]=a[t]||n}),n.p=a;var i=new XMLHttpRequest;i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback()}}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i=0;i<a.length;i++)r.apply(this,a[i]);function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var l=1;function v(e){var t;"auxclick"===e.type&&e.button!==l||((t=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],function e(t,n){if(!t||g<n)return!1;if(h(t))return!0;return e(t.parentNode,n+1)}(t,0))}function d(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}function m(e){var t=h(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}c.addEventListener("click",v),c.addEventListener("auxclick",v);var g=3;function w(e){if("auxclick"!==e.type||e.button===l){for(var t,n,r,a,i=e.target,o=0;o<=g&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;f(i)&&(t=i),h(i)&&(n=i),i=i.parentNode}n&&(a=m(n),t?(a.props.url=t.href,d(e,t,a)):vince(a.name,{props:a.props}))}}function h(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}c.addEventListener("submit",function(e){var t,n=e.target,r=m(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),c.addEventListener("click",w),c.addEventListener("auxclick",w)}();