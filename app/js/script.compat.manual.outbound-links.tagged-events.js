!function(){"use strict";var e,t,r=window.location,i=window.document,o=i.getElementById("plausible"),l=o.getAttribute("data-api")||(e=(e=o).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function a(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a={},n=(a.n=e,a.u=t&&t.u?t.u:r.href,a.d=o.getAttribute("data-domain"),a.r=i.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(a)),n.onreadystatechange=function(){4===n.readyState&&t&&t.callback&&t.callback({status:n.status})}}var n=window.plausible&&window.plausible.q||[];window.plausible=a;for(var s=0;s<n.length;s++)a.apply(this,n[s]);function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function f(e){var t,a;if("auxclick"!==e.type||e.button===c)return(t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||v<a)return!1;if(g(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(a=t)&&a.href&&a.host&&a.host!==r.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function d(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}function m(e){var e=g(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}i.addEventListener("click",f),i.addEventListener("auxclick",f);var v=3;function w(e){if("auxclick"!==e.type||e.button===c){for(var t,a,n,r,i=e.target,o=0;o<=v&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;p(i)&&(t=i),g(i)&&(a=i),i=i.parentNode}a&&(r=m(a),t?(r.props.url=t.href,d(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function g(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}i.addEventListener("submit",function(e){var t,a=e.target,n=m(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),i.addEventListener("click",w),i.addEventListener("auxclick",w)}();